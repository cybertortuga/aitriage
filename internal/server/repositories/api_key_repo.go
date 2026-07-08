package repositories

import (
	"database/sql"
	"time"
)

type APIKey struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	Status    string    `json:"status"`
	CreatedBy int       `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  *time.Time `json:"last_used"`
}

type APIKeyRepository struct {
	db *sql.DB
}

func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) List() ([]APIKey, error) {
	rows, err := r.db.Query(`
		SELECT id, name, prefix, status, created_by, created_at, last_used 
		FROM api_keys 
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		err := rows.Scan(&k.ID, &k.Name, &k.Prefix, &k.Status, &k.CreatedBy, &k.CreatedAt, &k.LastUsed)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *APIKeyRepository) Create(name, prefix, hash string, userID int) error {
	_, err := r.db.Exec(`
		INSERT INTO api_keys (name, prefix, key_hash, created_by)
		VALUES (?, ?, ?, ?)
	`, name, prefix, hash, userID)
	return err
}

func (r *APIKeyRepository) Revoke(id int) error {
	_, err := r.db.Exec("UPDATE api_keys SET status = 'REVOKED' WHERE id = ?", id)
	return err
}

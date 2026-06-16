package repositories

import (
	"context"
	"database/sql"
)

type ConfigRepository struct {
	db *sql.DB
}

func NewConfigRepository(db *sql.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

func (r *ConfigRepository) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT config_key, config_val FROM system_config")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	config := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		config[k] = v
	}
	return config, nil
}

func (r *ConfigRepository) Get(ctx context.Context, key string) (string, error) {
	var val string
	err := r.db.QueryRowContext(ctx, "SELECT config_val FROM system_config WHERE config_key = ?", key).Scan(&val)
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *ConfigRepository) Set(ctx context.Context, key, val string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO system_config (config_key, config_val)
		VALUES (?, ?)
		ON CONFLICT(config_key) DO UPDATE SET config_val = excluded.config_val, updated_at = CURRENT_TIMESTAMP
	`, key, val)
	return err
}

func (r *ConfigRepository) SetMany(ctx context.Context, config map[string]string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for k, v := range config {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO system_config (config_key, config_val)
			VALUES (?, ?)
			ON CONFLICT(config_key) DO UPDATE SET config_val = excluded.config_val, updated_at = CURRENT_TIMESTAMP
		`, k, v)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

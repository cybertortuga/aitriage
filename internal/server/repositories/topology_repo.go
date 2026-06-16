package repositories

import (
	"database/sql"
	"time"
)

type TopologyNode struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Risk      string    `json:"risk"`
	CreatedAt time.Time `json:"created_at"`
}

type TopologyLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type TopologyRepository struct {
	db *sql.DB
}

func NewTopologyRepository(db *sql.DB) *TopologyRepository {
	return &TopologyRepository{db: db}
}

func (r *TopologyRepository) List() ([]TopologyNode, error) {
	rows, err := r.db.Query("SELECT id, name, type, status, risk, created_at FROM topology_nodes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []TopologyNode
	for rows.Next() {
		var n TopologyNode
		err := rows.Scan(&n.ID, &n.Name, &n.Type, &n.Status, &n.Risk, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (r *TopologyRepository) Upsert(n TopologyNode) error {
	_, err := r.db.Exec(`
		INSERT INTO topology_nodes (id, name, type, status, risk)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status=excluded.status,
			risk=excluded.risk
	`, n.ID, n.Name, n.Type, n.Status, n.Risk)
	return err
}

func (r *TopologyRepository) ListLinks() ([]TopologyLink, error) {
	rows, err := r.db.Query("SELECT source, target FROM topology_links")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []TopologyLink
	for rows.Next() {
		var l TopologyLink
		err := rows.Scan(&l.Source, &l.Target)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, nil
}

func (r *TopologyRepository) UpsertLink(l TopologyLink) error {
	_, err := r.db.Exec(`
		INSERT INTO topology_links (source, target)
		VALUES (?, ?)
		ON CONFLICT(source, target) DO NOTHING
	`, l.Source, l.Target)
	return err
}

func (r *TopologyRepository) Clear() error {
	_, err := r.db.Exec("DELETE FROM topology_links")
	if err != nil {
		return err
	}
	_, err = r.db.Exec("DELETE FROM topology_nodes")
	return err
}

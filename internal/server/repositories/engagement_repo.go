package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/models"
)

type EngagementRepository struct {
	db *sql.DB
}

func NewEngagementRepository(db *sql.DB) *EngagementRepository {
	return &EngagementRepository{db: db}
}

func (r *EngagementRepository) Create(ctx context.Context, e *models.Engagement) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO engagements (product_id, name, description, scan_path, engagement_type, status, target_start, target_end, triggered_by, build_id, branch, commit_hash, scanner_version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, e.ProductID, e.Name, e.Description, e.ScanPath, e.EngagementType, e.Status, e.TargetStart, e.TargetEnd, e.TriggeredBy, e.BuildID, e.Branch, e.CommitHash, e.ScannerVersion)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *EngagementRepository) GetByID(ctx context.Context, id int64) (*models.Engagement, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, product_id, name, description, scan_path, engagement_type, status, started_at, completed_at, target_start, target_end, triggered_by, build_id, branch, commit_hash, scanner_version
		FROM engagements WHERE id = ?
	`, id)

	var e models.Engagement
	err := row.Scan(&e.ID, &e.ProductID, &e.Name, &e.Description, &e.ScanPath, &e.EngagementType, &e.Status, &e.StartedAt, &e.CompletedAt, &e.TargetStart, &e.TargetEnd, &e.TriggeredBy, &e.BuildID, &e.Branch, &e.CommitHash, &e.ScannerVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("engagement not found")
		}
		return nil, err
	}
	return &e, nil
}

func (r *EngagementRepository) List(ctx context.Context, productID int64) ([]models.Engagement, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, product_id, name, description, scan_path, engagement_type, status, started_at, completed_at, target_start, target_end, triggered_by, build_id, branch, commit_hash, scanner_version
		FROM engagements
		WHERE product_id = ?
		ORDER BY started_at DESC
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var engagements []models.Engagement
	for rows.Next() {
		var e models.Engagement
		if err := rows.Scan(&e.ID, &e.ProductID, &e.Name, &e.Description, &e.ScanPath, &e.EngagementType, &e.Status, &e.StartedAt, &e.CompletedAt, &e.TargetStart, &e.TargetEnd, &e.TriggeredBy, &e.BuildID, &e.Branch, &e.CommitHash, &e.ScannerVersion); err != nil {
			return nil, err
		}
		engagements = append(engagements, e)
	}
	return engagements, nil
}

func (r *EngagementRepository) ListAll(ctx context.Context) ([]models.Engagement, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, product_id, name, description, scan_path, engagement_type, status, started_at, completed_at, target_start, target_end, triggered_by, build_id, branch, commit_hash, scanner_version
		FROM engagements
		ORDER BY started_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var engagements []models.Engagement
	for rows.Next() {
		var e models.Engagement
		if err := rows.Scan(&e.ID, &e.ProductID, &e.Name, &e.Description, &e.ScanPath, &e.EngagementType, &e.Status, &e.StartedAt, &e.CompletedAt, &e.TargetStart, &e.TargetEnd, &e.TriggeredBy, &e.BuildID, &e.Branch, &e.CommitHash, &e.ScannerVersion); err != nil {
			return nil, err
		}
		engagements = append(engagements, e)
	}
	return engagements, nil
}

func (r *EngagementRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	var err error
	if status == "completed" {
		_, err = r.db.ExecContext(ctx, `UPDATE engagements SET status = ?, completed_at = CURRENT_TIMESTAMP WHERE id = ?`, status, id)
	} else {
		_, err = r.db.ExecContext(ctx, `UPDATE engagements SET status = ? WHERE id = ?`, status, id)
	}
	return err
}

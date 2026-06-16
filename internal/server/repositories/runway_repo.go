package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/models"
)

type RunwayRepository struct {
	db *sql.DB
}

func NewRunwayRepository(db *sql.DB) *RunwayRepository {
	return &RunwayRepository{db: db}
}

func (r *RunwayRepository) Create(ctx context.Context, session *models.RunwaySession) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO runway_sessions (product_id, status, current_step, auto_mode, threat_model, security_plan, remediation, poc, audit_report, scan_count_before, scan_count_after, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, session.ProductID, session.Status, session.CurrentStep, session.AutoMode, session.ThreatModel, session.SecurityPlan, session.Remediation, session.PoC, session.AuditReport, session.ScanCountBefore, session.ScanCountAfter, session.ErrorMessage)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	session.ID = id
	return nil
}

func (r *RunwayRepository) GetByID(ctx context.Context, id int64) (*models.RunwaySession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, product_id, status, current_step, auto_mode, threat_model, security_plan, remediation, poc, audit_report, scan_count_before, scan_count_after, error_message, created_at, updated_at
		FROM runway_sessions WHERE id = ?
	`, id)

	var s models.RunwaySession
	err := row.Scan(&s.ID, &s.ProductID, &s.Status, &s.CurrentStep, &s.AutoMode, &s.ThreatModel, &s.SecurityPlan, &s.Remediation, &s.PoC, &s.AuditReport, &s.ScanCountBefore, &s.ScanCountAfter, &s.ErrorMessage, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("runway session not found")
		}
		return nil, err
	}
	return &s, nil
}

func (r *RunwayRepository) GetActiveByProductID(ctx context.Context, productID int64) (*models.RunwaySession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, product_id, status, current_step, auto_mode, threat_model, security_plan, remediation, poc, audit_report, scan_count_before, scan_count_after, error_message, created_at, updated_at
		FROM runway_sessions
		WHERE product_id = ? AND status = 'in_progress'
		ORDER BY created_at DESC LIMIT 1
	`, productID)

	var s models.RunwaySession
	err := row.Scan(&s.ID, &s.ProductID, &s.Status, &s.CurrentStep, &s.AutoMode, &s.ThreatModel, &s.SecurityPlan, &s.Remediation, &s.PoC, &s.AuditReport, &s.ScanCountBefore, &s.ScanCountAfter, &s.ErrorMessage, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active runway session found")
		}
		return nil, err
	}
	return &s, nil
}

func (r *RunwayRepository) Update(ctx context.Context, session *models.RunwaySession) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE runway_sessions
		SET product_id = ?, status = ?, current_step = ?, auto_mode = ?, threat_model = ?, security_plan = ?, remediation = ?, poc = ?, audit_report = ?, scan_count_before = ?, scan_count_after = ?, error_message = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, session.ProductID, session.Status, session.CurrentStep, session.AutoMode, session.ThreatModel, session.SecurityPlan, session.Remediation, session.PoC, session.AuditReport, session.ScanCountBefore, session.ScanCountAfter, session.ErrorMessage, session.ID)
	return err
}

func (r *RunwayRepository) ListByProductID(ctx context.Context, productID int64) ([]models.RunwaySession, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, product_id, status, current_step, auto_mode, threat_model, security_plan, remediation, poc, audit_report, scan_count_before, scan_count_after, error_message, created_at, updated_at
		FROM runway_sessions
		WHERE product_id = ?
		ORDER BY created_at DESC
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.RunwaySession
	for rows.Next() {
		var s models.RunwaySession
		if err := rows.Scan(&s.ID, &s.ProductID, &s.Status, &s.CurrentStep, &s.AutoMode, &s.ThreatModel, &s.SecurityPlan, &s.Remediation, &s.PoC, &s.AuditReport, &s.ScanCountBefore, &s.ScanCountAfter, &s.ErrorMessage, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *RunwayRepository) ListAll(ctx context.Context) ([]models.RunwaySession, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, product_id, status, current_step, auto_mode, threat_model, security_plan, remediation, poc, audit_report, scan_count_before, scan_count_after, error_message, created_at, updated_at
		FROM runway_sessions
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.RunwaySession
	for rows.Next() {
		var s models.RunwaySession
		if err := rows.Scan(&s.ID, &s.ProductID, &s.Status, &s.CurrentStep, &s.AutoMode, &s.ThreatModel, &s.SecurityPlan, &s.Remediation, &s.PoC, &s.AuditReport, &s.ScanCountBefore, &s.ScanCountAfter, &s.ErrorMessage, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *RunwayRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM runway_sessions WHERE id = ?`, id)
	return err
}


package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/models"
)

type FindingRepository struct {
	db *sql.DB
}

func NewFindingRepository(db *sql.DB) *FindingRepository {
	return &FindingRepository{db: db}
}

func (r *FindingRepository) getIgnoredStatus(ctx context.Context, codeSnippet *string, defaultStatus string) string {
	if codeSnippet == nil || *codeSnippet == "" {
		return defaultStatus
	}
	hash := ContentHash(*codeSnippet)
	var reason string
	err := r.db.QueryRowContext(ctx, "SELECT reason FROM ignored_findings WHERE content_hash = ? LIMIT 1", hash).Scan(&reason)
	if err == nil {
		switch reason {
		case "False Positive":
			return "false_positive"
		case "Accepted Risk", "Won't Fix":
			return "risk_accepted"
		}
	}
	return defaultStatus
}

func (r *FindingRepository) Create(ctx context.Context, f *models.Finding) (int64, error) {
	if f.Status == "" {
		f.Status = "open"
	}
	status := r.getIgnoredStatus(ctx, f.CodeSnippet, f.Status)
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO findings (engagement_id, product_id, rule_id, title, severity, cvss_score, cve_id, cwe_id, file_path, line_number, col_number, code_snippet, description, impact, fix_suggestion, references_, hash_code, status, kanban_column, stack)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, f.EngagementID, f.ProductID, f.RuleID, f.Title, f.Severity, f.CVSSScore, f.CVEID, f.CWEID, f.FilePath, f.LineNumber, f.ColNumber, f.CodeSnippet, f.Description, f.Impact, f.FixSuggestion, f.References, f.HashCode, status, f.KanbanColumn, f.Stack)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *FindingRepository) BulkCreate(ctx context.Context, findings []models.Finding) error {
	if len(findings) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO findings (engagement_id, product_id, rule_id, title, severity, cvss_score, cve_id, cwe_id, file_path, line_number, col_number, code_snippet, description, impact, fix_suggestion, references_, hash_code, status, kanban_column, stack)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, f := range findings {
		if f.Status == "" {
			f.Status = "open"
		}
		status := r.getIgnoredStatus(ctx, f.CodeSnippet, f.Status)
		_, err := stmt.ExecContext(ctx, f.EngagementID, f.ProductID, f.RuleID, f.Title, f.Severity, f.CVSSScore, f.CVEID, f.CWEID, f.FilePath, f.LineNumber, f.ColNumber, f.CodeSnippet, f.Description, f.Impact, f.FixSuggestion, f.References, f.HashCode, status, f.KanbanColumn, f.Stack)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *FindingRepository) Deduplicate(ctx context.Context, productID int64) error {
	// Simple deduplication logic marking older identical hash_codes as duplicates
	_, err := r.db.ExecContext(ctx, `
		UPDATE findings 
		SET is_duplicate = 1, duplicate_of = (
			SELECT MIN(f2.id) FROM findings f2 
			WHERE f2.hash_code = findings.hash_code 
			AND f2.product_id = findings.product_id 
			AND f2.id != findings.id
		)
		WHERE product_id = ? 
		AND id != (
			SELECT MIN(f3.id) FROM findings f3 
			WHERE f3.hash_code = findings.hash_code 
			AND f3.product_id = findings.product_id
		)
	`, productID)
	return err
}

func (r *FindingRepository) GetByID(ctx context.Context, id int64) (*models.Finding, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, engagement_id, product_id, rule_id, title, severity, cvss_score, cve_id, cwe_id, file_path, line_number, col_number, code_snippet, description, impact, fix_suggestion, references_, hash_code, is_duplicate, duplicate_of, status, kanban_column, sla_deadline, sla_breached, risk_accepted, risk_accepted_by, risk_accepted_reason, risk_accepted_expiry, assigned_to, is_verified, verified_by, verified_at, created_at, updated_at, resolved_at, resolved_by, is_false_positive, fp_reason, stack, ai_triage_status, ai_triage_summary
		FROM findings WHERE id = ?
	`, id)

	var f models.Finding
	err := row.Scan(&f.ID, &f.EngagementID, &f.ProductID, &f.RuleID, &f.Title, &f.Severity, &f.CVSSScore, &f.CVEID, &f.CWEID, &f.FilePath, &f.LineNumber, &f.ColNumber, &f.CodeSnippet, &f.Description, &f.Impact, &f.FixSuggestion, &f.References, &f.HashCode, &f.IsDuplicate, &f.DuplicateOf, &f.Status, &f.KanbanColumn, &f.SLADeadline, &f.SLABreached, &f.RiskAccepted, &f.RiskAcceptedBy, &f.RiskAcceptedReason, &f.RiskAcceptedExpiry, &f.AssignedTo, &f.IsVerified, &f.VerifiedBy, &f.VerifiedAt, &f.CreatedAt, &f.UpdatedAt, &f.ResolvedAt, &f.ResolvedBy, &f.IsFalsePositive, &f.FPReason, &f.Stack, &f.AITriageStatus, &f.AITriageSummary)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("finding not found")
		}
		return nil, err
	}
	return &f, nil
}

func (r *FindingRepository) List(ctx context.Context, engagementID int64) ([]models.Finding, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, engagement_id, product_id, rule_id, title, severity, cvss_score, cve_id, cwe_id, file_path, line_number, col_number, code_snippet, description, impact, fix_suggestion, references_, hash_code, is_duplicate, duplicate_of, status, kanban_column, sla_deadline, sla_breached, risk_accepted, risk_accepted_by, risk_accepted_reason, risk_accepted_expiry, assigned_to, is_verified, verified_by, verified_at, created_at, updated_at, resolved_at, resolved_by, is_false_positive, fp_reason, stack, ai_triage_status, ai_triage_summary
		FROM findings
		WHERE engagement_id = ?
		ORDER BY created_at DESC
	`, engagementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []models.Finding
	for rows.Next() {
		var f models.Finding
		if err := rows.Scan(&f.ID, &f.EngagementID, &f.ProductID, &f.RuleID, &f.Title, &f.Severity, &f.CVSSScore, &f.CVEID, &f.CWEID, &f.FilePath, &f.LineNumber, &f.ColNumber, &f.CodeSnippet, &f.Description, &f.Impact, &f.FixSuggestion, &f.References, &f.HashCode, &f.IsDuplicate, &f.DuplicateOf, &f.Status, &f.KanbanColumn, &f.SLADeadline, &f.SLABreached, &f.RiskAccepted, &f.RiskAcceptedBy, &f.RiskAcceptedReason, &f.RiskAcceptedExpiry, &f.AssignedTo, &f.IsVerified, &f.VerifiedBy, &f.VerifiedAt, &f.CreatedAt, &f.UpdatedAt, &f.ResolvedAt, &f.ResolvedBy, &f.IsFalsePositive, &f.FPReason, &f.Stack, &f.AITriageStatus, &f.AITriageSummary); err != nil {
			return nil, err
		}
		findings = append(findings, f)
	}
	// Normalize: ensure no finding has an empty status
	for i := range findings {
		if findings[i].Status == "" {
			findings[i].Status = "open"
		}
	}
	return findings, nil
}

func (r *FindingRepository) ListAll(ctx context.Context) ([]models.Finding, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, engagement_id, product_id, rule_id, title, severity, cvss_score, cve_id, cwe_id, file_path, line_number, col_number, code_snippet, description, impact, fix_suggestion, references_, hash_code, is_duplicate, duplicate_of, status, kanban_column, sla_deadline, sla_breached, risk_accepted, risk_accepted_by, risk_accepted_reason, risk_accepted_expiry, assigned_to, is_verified, verified_by, verified_at, created_at, updated_at, resolved_at, resolved_by, is_false_positive, fp_reason, stack, ai_triage_status, ai_triage_summary
		FROM findings
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []models.Finding
	for rows.Next() {
		var f models.Finding
		if err := rows.Scan(&f.ID, &f.EngagementID, &f.ProductID, &f.RuleID, &f.Title, &f.Severity, &f.CVSSScore, &f.CVEID, &f.CWEID, &f.FilePath, &f.LineNumber, &f.ColNumber, &f.CodeSnippet, &f.Description, &f.Impact, &f.FixSuggestion, &f.References, &f.HashCode, &f.IsDuplicate, &f.DuplicateOf, &f.Status, &f.KanbanColumn, &f.SLADeadline, &f.SLABreached, &f.RiskAccepted, &f.RiskAcceptedBy, &f.RiskAcceptedReason, &f.RiskAcceptedExpiry, &f.AssignedTo, &f.IsVerified, &f.VerifiedBy, &f.VerifiedAt, &f.CreatedAt, &f.UpdatedAt, &f.ResolvedAt, &f.ResolvedBy, &f.IsFalsePositive, &f.FPReason, &f.Stack, &f.AITriageStatus, &f.AITriageSummary); err != nil {
			return nil, err
		}
		findings = append(findings, f)
	}
	// Normalize: ensure no finding has an empty status
	for i := range findings {
		if findings[i].Status == "" {
			findings[i].Status = "open"
		}
	}
	return findings, nil
}

// EnsureSLA evaluates Findings and marks those past deadline as sla_breached
func (r *FindingRepository) EnsureSLA(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE findings 
		SET sla_breached = 1 
		WHERE sla_deadline IS NOT NULL 
		AND status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive') 
		AND CURRENT_TIMESTAMP > sla_deadline
	`)
	return err
}

func (r *FindingRepository) UpdateKanbanColumn(ctx context.Context, id int64, column string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE findings SET kanban_column = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, column, id)
	return err
}

func (r *FindingRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE findings SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, status, id)
	return err
}

func (r *FindingRepository) UpdateAITriage(ctx context.Context, id int64, status, summary string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE findings SET ai_triage_status = ?, ai_triage_summary = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, status, summary, id)
	return err
}

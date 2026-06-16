package repositories

import (
	"database/sql"
	"time"
)

type Report struct {
	ID          int       `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	TargetScope string    `json:"target_scope"`
	Format      string    `json:"format"`
	Status      string    `json:"status"`
	DownloadURL string    `json:"download_url"`
}

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) ListReports() ([]Report, error) {
	rows, err := r.db.Query(`
		SELECT id, timestamp, target_scope, format, status, COALESCE(download_url, '')
		FROM reports
		ORDER BY timestamp DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []Report
	for rows.Next() {
		var rep Report
		if err := rows.Scan(&rep.ID, &rep.Timestamp, &rep.TargetScope, &rep.Format, &rep.Status, &rep.DownloadURL); err != nil {
			return nil, err
		}
		reports = append(reports, rep)
	}
	return reports, nil
}

func (r *ReportRepository) CreateReport(scope, format, status string) error {
	_, err := r.db.Exec(`
		INSERT INTO reports (target_scope, format, status)
		VALUES (?, ?, ?)
	`, scope, format, status)
	return err
}

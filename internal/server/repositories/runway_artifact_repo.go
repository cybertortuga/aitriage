package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/models"
)

type RunwayArtifactRepository struct {
	db *sql.DB
}

func NewRunwayArtifactRepository(db *sql.DB) *RunwayArtifactRepository {
	return &RunwayArtifactRepository{db: db}
}

func (r *RunwayArtifactRepository) UpsertMany(ctx context.Context, artifacts []models.RunwayArtifact) error {
	if len(artifacts) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin runway artifact transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO runway_artifacts (session_id, kind, media_type, schema_version, content, sha256)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(session_id, kind) DO UPDATE SET
			media_type = excluded.media_type,
			schema_version = excluded.schema_version,
			content = excluded.content,
			sha256 = excluded.sha256,
			created_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("prepare runway artifact upsert: %w", err)
	}
	defer stmt.Close()

	for _, artifact := range artifacts {
		if !models.IsValidRunwayArtifactKind(artifact.Kind) {
			return fmt.Errorf("unsupported runway artifact kind %q", artifact.Kind)
		}
		if _, err := stmt.ExecContext(ctx, artifact.SessionID, artifact.Kind, artifact.MediaType, artifact.SchemaVersion, artifact.Content, artifact.SHA256); err != nil {
			return fmt.Errorf("upsert runway artifact %s: %w", artifact.Kind, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit runway artifacts: %w", err)
	}
	return nil
}

func (r *RunwayArtifactRepository) Get(ctx context.Context, sessionID int64, kind string) (*models.RunwayArtifact, error) {
	if !models.IsValidRunwayArtifactKind(kind) {
		return nil, fmt.Errorf("unsupported runway artifact kind %q", kind)
	}
	row := r.db.QueryRowContext(ctx, `
		SELECT id, session_id, kind, media_type, schema_version, content, sha256, length(content), created_at
		FROM runway_artifacts
		WHERE session_id = ? AND kind = ?
	`, sessionID, kind)
	var artifact models.RunwayArtifact
	if err := row.Scan(&artifact.ID, &artifact.SessionID, &artifact.Kind, &artifact.MediaType, &artifact.SchemaVersion, &artifact.Content, &artifact.SHA256, &artifact.SizeBytes, &artifact.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get runway artifact: %w", err)
	}
	return &artifact, nil
}

func (r *RunwayArtifactRepository) ListMetadata(ctx context.Context, sessionID int64) ([]models.RunwayArtifact, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, session_id, kind, media_type, schema_version, sha256, length(content), created_at
		FROM runway_artifacts
		WHERE session_id = ?
		ORDER BY id
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list runway artifacts: %w", err)
	}
	defer rows.Close()

	artifacts := make([]models.RunwayArtifact, 0, 6)
	for rows.Next() {
		var artifact models.RunwayArtifact
		if err := rows.Scan(&artifact.ID, &artifact.SessionID, &artifact.Kind, &artifact.MediaType, &artifact.SchemaVersion, &artifact.SHA256, &artifact.SizeBytes, &artifact.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan runway artifact metadata: %w", err)
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, rows.Err()
}

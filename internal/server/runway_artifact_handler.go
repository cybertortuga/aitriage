package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dodobrands/aitriage/internal/agent/graph"
	"github.com/dodobrands/aitriage/internal/models"
	"github.com/dodobrands/aitriage/internal/server/middleware"
)

const maxRunwayArtifactBytes = 25 << 20

func (s *Server) persistRunwayArtifacts(ctx context.Context, sessionID int64, state *graph.AgentState) error {
	if state == nil {
		return fmt.Errorf("persist runway artifacts: nil agent state")
	}
	handoff := state.AgentHandoff
	if handoff == nil {
		built, err := graph.BuildAgentHandoff(state, time.Now().UTC())
		if err != nil {
			return err
		}
		handoff = &built
	}

	agentDataJSON, err := json.MarshalIndent(handoff.AgentData, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal agent data: %w", err)
	}
	triage, err := graph.BuildTriageFindingsArtifact(state)
	if err != nil {
		return err
	}
	triageJSON, err := json.MarshalIndent(triage, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal triage findings artifact: %w", err)
	}

	artifacts := []models.RunwayArtifact{
		newRunwayArtifact(sessionID, models.RunwayArtifactSummaryMarkdown, "text/markdown; charset=utf-8", handoff.SchemaVersion, handoff.SummaryMarkdown),
		newRunwayArtifact(sessionID, models.RunwayArtifactRemediationPromptMarkdown, "text/markdown; charset=utf-8", handoff.SchemaVersion, handoff.RemediationPromptMarkdown),
		newRunwayArtifact(sessionID, models.RunwayArtifactAgentDataJSON, "application/json; charset=utf-8", handoff.SchemaVersion, string(agentDataJSON)+"\n"),
		newRunwayArtifact(sessionID, models.RunwayArtifactReportMarkdown, "text/markdown; charset=utf-8", handoff.SchemaVersion, state.ReportMarkdown),
		newRunwayArtifact(sessionID, models.RunwayArtifactFixSpecMarkdown, "text/markdown; charset=utf-8", handoff.SchemaVersion, state.AIFixSpec),
		newRunwayArtifact(sessionID, models.RunwayArtifactTriageFindingsJSON, "application/json; charset=utf-8", graph.TriageArtifactSchemaVersion, string(triageJSON)+"\n"),
	}
	for _, artifact := range artifacts {
		if len(artifact.Content) > maxRunwayArtifactBytes {
			return fmt.Errorf("runway artifact %s exceeds %d bytes", artifact.Kind, maxRunwayArtifactBytes)
		}
	}
	return s.runwayArtifactRepo.UpsertMany(ctx, artifacts)
}

func newRunwayArtifact(sessionID int64, kind, mediaType string, schemaVersion int, content string) models.RunwayArtifact {
	digest := sha256.Sum256([]byte(content))
	return models.RunwayArtifact{
		SessionID:     sessionID,
		Kind:          kind,
		MediaType:     mediaType,
		SchemaVersion: schemaVersion,
		Content:       content,
		SHA256:        hex.EncodeToString(digest[:]),
		SizeBytes:     len(content),
	}
}

func (s *Server) handleRunwayArtifactManifest(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := runwayPathID(w, r)
	if !ok {
		return
	}
	if _, err := s.runwayRepo.GetByID(r.Context(), sessionID); err != nil {
		jsonError(w, "runway session not found", http.StatusNotFound)
		return
	}
	artifacts, err := s.runwayArtifactRepo.ListMetadata(r.Context(), sessionID)
	if err != nil {
		jsonError(w, "failed to load runway artifacts", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":         true,
		"session_id": sessionID,
		"artifacts":  artifacts,
	})
}

func (s *Server) handleRunwayHandoff(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := runwayPathID(w, r)
	if !ok {
		return
	}
	session, err := s.runwayRepo.GetByID(r.Context(), sessionID)
	if err != nil {
		jsonError(w, "runway session not found", http.StatusNotFound)
		return
	}
	prompt, err := s.runwayArtifactRepo.Get(r.Context(), sessionID, models.RunwayArtifactRemediationPromptMarkdown)
	if err != nil {
		jsonError(w, "failed to load remediation prompt", http.StatusInternalServerError)
		return
	}
	agentDataArtifact, err := s.runwayArtifactRepo.Get(r.Context(), sessionID, models.RunwayArtifactAgentDataJSON)
	if err != nil {
		jsonError(w, "failed to load agent data", http.StatusInternalServerError)
		return
	}
	if prompt == nil || agentDataArtifact == nil {
		jsonError(w, "agent handoff is not available for this session", http.StatusNotFound)
		return
	}
	var agentData graph.AIAgentData
	if err := json.Unmarshal([]byte(agentDataArtifact.Content), &agentData); err != nil {
		jsonError(w, "stored agent data is invalid", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":                            true,
		"schema_version":                prompt.SchemaVersion,
		"session_id":                    sessionID,
		"generated_at":                  prompt.CreatedAt,
		"session_status":                session.Status,
		"gate_status":                   agentData.GateStatus,
		"remediation_prompt_markdown":   prompt.Content,
		"remediation_prompt_sha256":     prompt.SHA256,
		"remediation_prompt_size_bytes": prompt.SizeBytes,
		"agent_data":                    agentData,
		"agent_data_sha256":             agentDataArtifact.SHA256,
		"agent_data_size_bytes":         agentDataArtifact.SizeBytes,
	})
}

func (s *Server) handleRunwayArtifactDownload(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := runwayPathID(w, r)
	if !ok {
		return
	}
	kind := strings.TrimSpace(r.PathValue("kind"))
	if !models.IsValidRunwayArtifactKind(kind) {
		jsonError(w, "unsupported artifact kind", http.StatusBadRequest)
		return
	}
	if kind == models.RunwayArtifactTriageFindingsJSON {
		claims, _ := middleware.ExtractClaims(r)
		role := strings.ToLower(claims.GlobalRole)
		if role != "superadmin" && role != "admin" && role != "manager" {
			jsonError(w, "insufficient permission for canonical triage artifact", http.StatusForbidden)
			return
		}
	}
	artifact, err := s.runwayArtifactRepo.Get(r.Context(), sessionID, kind)
	if err != nil {
		jsonError(w, "failed to load runway artifact", http.StatusInternalServerError)
		return
	}
	if artifact == nil {
		jsonError(w, "runway artifact not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", artifact.MediaType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, runwayArtifactFilename(sessionID, kind)))
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("ETag", `"`+artifact.SHA256+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(artifact.Content))
}

func runwayPathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		jsonError(w, "invalid session id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func runwayArtifactFilename(sessionID int64, kind string) string {
	extension := ".md"
	if strings.HasSuffix(kind, "_json") {
		extension = ".json"
	}
	return fmt.Sprintf("runway-%d-%s%s", sessionID, strings.TrimSuffix(kind, "_markdown"), extension)
}

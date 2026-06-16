package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cybertortuga/aitriage/internal/server/repositories"
	"github.com/cybertortuga/aitriage/internal/server/utils"
)

type ReportHandler struct {
	findingRepo    *repositories.FindingRepository
	engagementRepo *repositories.EngagementRepository
	productRepo    *repositories.ProductRepository
	reportRepo     *repositories.ReportRepository
}

func NewReportHandler(findingRepo *repositories.FindingRepository, engagementRepo *repositories.EngagementRepository, productRepo *repositories.ProductRepository, reportRepo *repositories.ReportRepository) *ReportHandler {
	return &ReportHandler{
		findingRepo:    findingRepo,
		engagementRepo: engagementRepo,
		productRepo:    productRepo,
		reportRepo:     reportRepo,
	}
}

func (h *ReportHandler) HandleExecutiveReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	findings, err := h.findingRepo.ListAll(ctx)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	summary := struct {
		TotalFindings int            `json:"total_findings"`
		BySeverity    map[string]int `json:"by_severity"`
		ByStatus      map[string]int `json:"by_status"`
	}{
		TotalFindings: len(findings),
		BySeverity:    make(map[string]int),
		ByStatus:      make(map[string]int),
	}

	for _, f := range findings {
		summary.BySeverity[f.Severity]++
		summary.ByStatus[f.Status]++
	}

	format := r.URL.Query().Get("format")
	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment;filename=executive_report.csv")
		writer := csv.NewWriter(w)
		writer.Write([]string{"Severity", "Count"})
		for sev, count := range summary.BySeverity {
			writer.Write([]string{sev, strconv.Itoa(count)})
		}
		writer.Flush()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *ReportHandler) HandleEngagementReport(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/reports/engagement/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		utils.JSONError(w, "engagement_id is required", http.StatusBadRequest)
		return
	}

	engagementID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		utils.JSONError(w, "invalid engagement_id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	engagement, err := h.engagementRepo.GetByID(ctx, engagementID)
	if err != nil {
		utils.JSONError(w, "Engagement not found", http.StatusNotFound)
		return
	}

	findings, err := h.findingRepo.List(ctx, engagementID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	report := struct {
		EngagementName string `json:"engagement_name"`
		Findings       []any  `json:"findings"`
	}{
		EngagementName: engagement.Name,
	}

	for _, f := range findings {
		filePath := ""
		if f.FilePath != nil {
			filePath = *f.FilePath
		}
		lineNum := ""
		if f.LineNumber != nil {
			lineNum = strconv.Itoa(*f.LineNumber)
		}

		report.Findings = append(report.Findings, map[string]any{
			"title":    f.Title,
			"severity": f.Severity,
			"file":     filePath,
			"line":     lineNum,
			"status":   f.Status,
		})
	}

	format := r.URL.Query().Get("format")
	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=engagement_%d_report.csv", engagementID))
		writer := csv.NewWriter(w)
		writer.Write([]string{"Title", "Severity", "File", "Line", "Status"})
		for _, f := range findings {
			filePath := ""
			if f.FilePath != nil {
				filePath = *f.FilePath
			}
			lineNum := ""
			if f.LineNumber != nil {
				lineNum = strconv.Itoa(*f.LineNumber)
			}
			writer.Write([]string{f.Title, f.Severity, filePath, lineNum, f.Status})
		}
		writer.Flush()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *ReportHandler) HandleListReportHistory(w http.ResponseWriter, r *http.Request) {
	reports, err := h.reportRepo.ListReports()
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"reports": reports,
	})
}

func (h *ReportHandler) HandleGenerateReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Format  string `json:"format"`
		Scope   string `json:"scope"`
		Options struct {
			IncludeDeps bool `json:"include_deps"`
			Sign        bool `json:"sign"`
		} `json:"options"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Format == "" {
		req.Format = "SARIF"
	}
	if req.Scope == "" {
		req.Scope = "all-findings"
	}

	if err := h.reportRepo.CreateReport(req.Scope, req.Format, "READY"); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"format": req.Format,
		"scope":  req.Scope,
	})
}

package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/cybertortuga/aitriage/internal/config"
	"github.com/cybertortuga/aitriage/internal/engine"
	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/report/scorer"
	"github.com/cybertortuga/aitriage/internal/scanner/deps"
	"github.com/cybertortuga/aitriage/internal/scanner/detector"
)

type ScanOptions struct {
	ForceStack    string
	UniversalOnly bool
	FileFilter    []string // If non-empty, scan ONLY these files (absolute paths)
}

type ScanReport struct {
	ProjectPath         string               `json:"project_path"`
	Stacks              []detector.Stack     `json:"stacks"`
	Results             []core.CheckResult   `json:"results"`
	HasCriticalFailures bool                 `json:"has_critical_failures"`
	SecurityScore       int                  `json:"security_score"`
	SecurityGrade       string               `json:"security_grade"`
	Dependencies        []deps.Dependency    `json:"dependencies"`
	DependencyGraph     deps.DependencyGraph `json:"dependency_graph"`
	AISummary           string               `json:"ai_summary,omitempty"`
	TotalFiles          int                  `json:"total_files"`
	RulesApplied        int                  `json:"rules_applied"`
	ScanDuration        time.Duration        `json:"scan_duration_ms"`
	Config              *config.Config       `json:"-"` // not serialized; used by CLI
}

// ToJSON сериализует ScanReport в JSON. Используется MCP tools и agent mode.
func (r ScanReport) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// Scan выполняет полное детерминированное сканирование проекта.
// Принимает context для корректной отмены из MCP и agent режимов.
// Возвращает ScanReport и ошибку — НЕ вызывает os.Exit.
func Scan(ctx context.Context, projectPath string, opts ScanOptions) (ScanReport, error) {
	start := time.Now()

	empty := ScanReport{
		ProjectPath: projectPath,
		Stacks:      []detector.Stack{detector.UnknownStack},
		Results:     nil,
	}

	if info, err := os.Stat(projectPath); err != nil || !info.IsDir() {
		return empty, fmt.Errorf("path %q is not a valid directory", projectPath)
	}

	// Проверить отмену контекста
	select {
	case <-ctx.Done():
		return empty, ctx.Err()
	default:
	}

	ws, err := core.NewWorkspace(projectPath)
	if err != nil {
		slog.Error("Failed to create workspace", "error", err)
		return empty, fmt.Errorf("failed to create workspace: %w", err)
	}

	projects := detector.DetectProjects(ws)
	ws.Projects = projects

	// Apply file filter (for --diff / --staged) — keep only matching files
	if len(opts.FileFilter) > 0 {
		allowed := make(map[string]bool, len(opts.FileFilter))
		for _, f := range opts.FileFilter {
			abs, err := filepath.Abs(f)
			if err == nil {
				allowed[abs] = true
			} else {
				allowed[f] = true
			}
		}
		for _, proj := range projects {
			filtered := proj.Files[:0]
			for _, fi := range proj.Files {
				if allowed[fi.Path] {
					filtered = append(filtered, fi)
				}
			}
			proj.Files = filtered
		}
	}

	var allStacks []detector.Stack
	seenStacks := make(map[detector.Stack]bool)
	for _, p := range projects {
		st := detector.Stack(p.Stack)
		if !seenStacks[st] {
			allStacks = append(allStacks, st)
			seenStacks[st] = true
		}
	}

	var results []core.CheckResult

	// Init YAML Rule Engine
	eng, err := engine.NewEngine(ws.Config)
	if err != nil {
		slog.Error("Failed to initialize rule engine", "error", err)
		return empty, fmt.Errorf("failed to initialize rule engine: %w", err)
	}

	// Run engine on each project in the workspace
	for _, proj := range projects {
		// Проверить отмену контекста между проектами
		select {
		case <-ctx.Done():
			return empty, ctx.Err()
		default:
		}

		if opts.UniversalOnly && proj.Stack != string(detector.Universal) && proj.Stack != string(detector.UnknownStack) {
			continue // skip specific stacks if forced to universal only
		}

		if opts.ForceStack != "" && proj.Stack != opts.ForceStack {
			proj.Stack = opts.ForceStack
		}

		res := eng.Run(proj)
		results = append(results, res...)
	}

	// Deduplicate: project-level rules (no file) — keep one per ID
	// File-level rules — keep one per ID+File+Line combination
	seen := make(map[string]bool)
	deduped := results[:0]
	for _, r := range results {
		var key string
		if r.File == "" {
			key = r.ID // project-level: one per rule
		} else {
			key = fmt.Sprintf("%s|%s|%d", r.ID, r.File, r.Line)
		}
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, r)
		}
	}
	results = deduped

	// Apply Audit Statuses
	auditStore := core.NewAuditStore(projectPath)
	for i := range results {
		relPath := results[i].File
		// Normalize to relative path for consistent audit keys
		if relPath != "" && filepath.IsAbs(relPath) {
			relPath, _ = filepath.Rel(projectPath, relPath)
		}
		status := auditStore.GetStatus(results[i].ID, relPath)
		results[i].AuditStatus = status
	}

	hasCritical, securityScore := scorer.Calculate(results)
	securityGrade := "A+"
	switch {
	case securityScore < 50:
		securityGrade = "F"
	case securityScore < 65:
		securityGrade = "D"
	case securityScore < 80:
		securityGrade = "C"
	case securityScore < 90:
		securityGrade = "B"
	case securityScore < 100:
		securityGrade = "A"
	}

	for i := range results {
		results[i].OWASPMapping = scorer.GetOWASP(results[i].ID)
	}

	projectGraph := deps.GenerateGraph(ws)

	return ScanReport{
		ProjectPath:         projectPath,
		Stacks:              allStacks,
		Results:             results,
		HasCriticalFailures: hasCritical,
		SecurityScore:       securityScore,
		SecurityGrade:       securityGrade,
		Dependencies:        projectGraph.Nodes,
		DependencyGraph:     projectGraph,
		TotalFiles:          len(ws.Files),
		RulesApplied:        len(results),
		ScanDuration:        time.Since(start),
		Config:              ws.Config,
	}, nil
}

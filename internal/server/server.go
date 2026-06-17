package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cybertortuga/aitriage/internal/agent/architect"
	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/config"
	"github.com/cybertortuga/aitriage/internal/engine"
	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/engine/orchestrator"
	"github.com/cybertortuga/aitriage/internal/models"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
	"github.com/cybertortuga/aitriage/internal/scanner/deps"
	"github.com/cybertortuga/aitriage/internal/scanner/external"
	"github.com/cybertortuga/aitriage/internal/server/handlers"
	"github.com/cybertortuga/aitriage/internal/server/middleware"
	"github.com/cybertortuga/aitriage/internal/server/repositories"
	"golang.org/x/time/rate"
)

var globalLimiter = rate.NewLimiter(rate.Every(100*time.Millisecond), 10)

type Server struct {
	hostPrefix     string
	llmClient      llm.Client
	lastResult     *llm.RichScanResult
	userRepo       *repositories.UserRepository
	productRepo    *repositories.ProductRepository
	engagementRepo *repositories.EngagementRepository
	findingRepo    *repositories.FindingRepository
	auditRepo      *repositories.AuditRepository
	notifRepo      *repositories.NotificationRepository
	metricsRepo    *repositories.MetricsRepository
	apiKeyRepo     *repositories.APIKeyRepository
	topologyRepo   *repositories.TopologyRepository
	configRepo     *repositories.ConfigRepository
	reportRepo     *repositories.ReportRepository
	chatRepo       *repositories.ChatRepository
	ignoreRepo     *repositories.IgnoreRepository
	runwayRepo     *repositories.RunwayRepository
	db             *sql.DB
	engine         *engine.Engine
}

func NewServer(hostPrefix string, db *sql.DB) *Server {
	userRepo := repositories.NewUserRepository(db)
	productRepo := repositories.NewProductRepository(db)
	eng, err := engine.NewEngine(nil)
	if err != nil {
		slog.Error("CRITICAL: Failed to initialize security engine", "error", err)
	} else {
		slog.Info("Security engine initialized", "rules_count", len(eng.Rules))
	}

	// Eagerly initialize LLM client from env vars so chat works without a scan.
	var llmCl llm.Client
	cfg := config.LoadConfig(".")
	if cfg.LLM.APIKey != "" {
		client, llmErr := llm.NewClient(llm.Config{
			Provider: cfg.LLM.Provider,
			Model:    cfg.LLM.Model,
			APIKey:   cfg.LLM.APIKey,
			BaseURL:  cfg.LLM.BaseURL,
			Timeout:  cfg.LLM.Timeout,
		})
		if llmErr == nil {
			llmCl = client
			slog.Info("LLM client initialized at startup", "provider", cfg.LLM.Provider)
		} else {
			slog.Warn("LLM client init failed", "error", llmErr)
		}
	}

	return &Server{
		hostPrefix:     hostPrefix,
		llmClient:      llmCl,
		userRepo:       userRepo,
		productRepo:    productRepo,
		engagementRepo: repositories.NewEngagementRepository(db),
		findingRepo:    repositories.NewFindingRepository(db),
		auditRepo:      repositories.NewAuditRepository(db),
		notifRepo:      repositories.NewNotificationRepository(db),
		metricsRepo:    repositories.NewMetricsRepository(db),
		apiKeyRepo:     repositories.NewAPIKeyRepository(db),
		topologyRepo:   repositories.NewTopologyRepository(db),
		configRepo:     repositories.NewConfigRepository(db),
		reportRepo:     repositories.NewReportRepository(db),
		chatRepo:       repositories.NewChatRepository(db),
		ignoreRepo:     repositories.NewIgnoreRepository(db),
		runwayRepo:     repositories.NewRunwayRepository(db),
		db:             db,
		engine:         eng,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/api/scan", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleScan)))
	mux.Handle("/api/browser", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(s.handleBrowser)))
	mux.Handle("/api/triage", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleTriage)))
	mux.Handle("/api/chat", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(s.handleChat)))
	mux.Handle("/api/rules", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(s.handleRules)))

	// Product Management
	productHandler := handlers.NewProductHandler(s.productRepo)
	mux.Handle("/api/products", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			productHandler.HandleListProducts(w, r)
		case "POST":
			productHandler.HandleCreateProduct(w, r)
		case "PUT":
			productHandler.HandleUpdateProduct(w, r)
		case "DELETE":
			productHandler.HandleDeleteProduct(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/products/types", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(productHandler.HandleListProductTypes)))
	mux.Handle("/api/products/members", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(productHandler.HandleAddProductMember)))

	// Engagement Tracking
	engagementHandler := handlers.NewEngagementHandler(s.engagementRepo)
	mux.Handle("/api/engagements", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			engagementHandler.HandleListEngagements(w, r)
		case "POST":
			engagementHandler.HandleCreateEngagement(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))

	// Findings Management
	findingHandler := handlers.NewFindingHandler(s.findingRepo)
	mux.Handle("/api/findings", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			findingHandler.HandleListFindings(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/findings/", middleware.PermissionMiddleware("admin", "manager", "developer")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/ai-triage") {
			if r.Method == "POST" {
				s.handleAITriage(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		} else if r.Method == "PUT" {
			findingHandler.HandleUpdateFinding(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))

	// User Management (Admin)
	authHandler := handlers.NewAuthHandler(s.userRepo)
	adminUsersHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			authHandler.HandleListUsers(w, r)
		case "POST":
			authHandler.HandleCreateUser(w, r)
		case "DELETE":
			authHandler.HandleDeleteUser(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// System Configuration (Admin)
	configHandler := handlers.NewConfigHandler(s.configRepo)
	mux.Handle("/api/admin/config", middleware.PermissionMiddleware("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			configHandler.HandleGetConfig(w, r)
		case "POST", "PUT":
			configHandler.HandleUpdateConfig(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/admin/users", middleware.PermissionMiddleware("admin")(adminUsersHandler))

	mux.HandleFunc("/api/login", authHandler.HandleLogin)
	mux.Handle("/api/me", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(authHandler.HandleMe)))

	// Notifications
	notifHandler := handlers.NewNotificationHandler(s.notifRepo)
	mux.Handle("/api/notifications", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(notifHandler.HandleListNotifications)))
	mux.Handle("/api/notifications/", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(notifHandler.HandleMarkAsRead)))

	// Audit
	auditHandler := handlers.NewAuditHandler(s.auditRepo)
	mux.Handle("/api/audit", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(auditHandler.HandleListAuditLogs)))

	// Metrics
	metricsHandler := handlers.NewMetricsHandler(s.metricsRepo)
	mux.Handle("/api/metrics", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(metricsHandler.HandleGetDashboardMetrics)))

	// Reports
	reportHandler := handlers.NewReportHandler(s.findingRepo, s.engagementRepo, s.productRepo, s.reportRepo)
	mux.Handle("/api/reports/executive", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(reportHandler.HandleExecutiveReport)))
	mux.Handle("/api/reports/engagement/", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(reportHandler.HandleEngagementReport)))
	mux.Handle("/api/reports/history", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(reportHandler.HandleListReportHistory)))
	mux.Handle("/api/reports/generate", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(reportHandler.HandleGenerateReport)))

	mux.Handle("/api/analyze", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleAnalyze)))
	mux.Handle("/api/file", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(s.handleFile)))

	topologyHandler := handlers.NewTopologyHandler(s.topologyRepo)
	mux.Handle("/api/topology", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(topologyHandler.HandleGetTopology)))

	apiKeyHandler := handlers.NewAPIKeyHandler(s.apiKeyRepo)
	mux.Handle("/api/admin/keys", middleware.PermissionMiddleware("admin")(http.HandlerFunc(apiKeyHandler.HandleListKeys)))
	mux.Handle("/api/admin/keys/create", middleware.PermissionMiddleware("admin")(http.HandlerFunc(apiKeyHandler.HandleCreateKey)))
	mux.Handle("/api/admin/keys/", middleware.PermissionMiddleware("admin")(http.HandlerFunc(apiKeyHandler.HandleRevokeKey)))

	mux.Handle("/api/admin/clear-cache", middleware.PermissionMiddleware("admin")(http.HandlerFunc(s.handleClearCache)))
	mux.Handle("/api/admin/purge", middleware.PermissionMiddleware("admin")(http.HandlerFunc(s.handlePurgeDatabase)))
	mux.Handle("/api/admin/rebuild", middleware.PermissionMiddleware("admin")(http.HandlerFunc(s.handleRebuild)))

	mux.HandleFunc("/api/health", s.handleHealth)
	mux.Handle("/api/ai-summary", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(s.handleSummary)))

	mux.Handle("/api/chat/sessions", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(s.handleChatSessions)))
	mux.Handle("/api/chat/sessions/", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(s.handleChatSession)))
	mux.Handle("/api/chat/messages", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(s.handleChatMessages)))

	// SecureCoder-Compatible API
	mux.HandleFunc("/api/securecoder/config", s.handleSecureCoderConfig)
	mux.Handle("/api/securecoder/scan", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleSecureCoderScan)))
	mux.Handle("/api/securecoder/scan-directory", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleSecureCoderScanDirectory)))
	mux.Handle("/api/securecoder/ignore", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleSecureCoderIgnore)))
	mux.Handle("/api/securecoder/ignored", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(s.handleSecureCoderIgnored)))
	mux.Handle("/api/securecoder/fix_completed", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleSecureCoderFixCompleted)))
	mux.Handle("/api/securecoder/dependency/scan", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleSecureCoderDepScan)))
	mux.Handle("/api/securecoder/wiz/status", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(s.handleWizStatus)))
	mux.Handle("/api/securecoder/wiz/login", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleWizLogin)))
	mux.Handle("/api/securecoder/wiz/login/poll", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(s.handleWizLoginPoll)))
	mux.Handle("/api/securecoder/wiz/logout", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleWizLogout)))
	mux.Handle("/api/securecoder/ignore-file", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleSecureCoderIgnoreFile)))

	mux.Handle("/api/runway", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleRunway)))
	mux.Handle("/api/runway/all", middleware.PermissionMiddleware("admin", "manager", "viewer")(http.HandlerFunc(s.handleRunwayAll)))
	mux.Handle("/api/runway/export/", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleRunwayExport)))
	mux.Handle("/api/runway/", middleware.PermissionMiddleware("admin", "manager")(http.HandlerFunc(s.handleRunwaySession)))

	// UI
	mux.HandleFunc("/", handleUI)

	handler := middleware.SecurityHeadersMiddleware(
		middleware.RateLimitMiddleware(
			mux,
		),
	)

	handler.ServeHTTP(w, r)
}

func (s *Server) limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !globalLimiter.Allow() {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── API Handlers ─────────────────────────────────────────────────────────────

type scanRequest struct {
	Path     string `json:"path"`
	Stack    string `json:"stack,omitempty"`
	External bool   `json:"external,omitempty"`
}

type findingDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Suggestion  string `json:"suggestion"`
	OWASP       string `json:"owasp,omitempty"`
	AuditStatus string `json:"audit_status"`
	Stack       string `json:"stack"`
}

type scanResponse struct {
	Ok            bool               `json:"ok"`
	ScanID        string             `json:"scan_id"`
	Findings      []findingDTO       `json:"findings"`
	Dependencies  []deps.Dependency  `json:"dependencies"`
	Stacks        []string           `json:"stacks"`
	SecurityScore int                `json:"security_score"`
	SecurityGrade string             `json:"security_grade"`
	HealthCheck   healthcheck.Result `json:"health_check"`
	Duration      string             `json:"duration"`
	Error         string             `json:"error,omitempty"`
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	var req scanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Path == "" {
		jsonError(w, "path is required", http.StatusBadRequest)
		return
	}

	containerPath := req.Path
	if s.hostPrefix != "" && !strings.HasPrefix(req.Path, s.hostPrefix) {
		containerPath = filepath.Join(s.hostPrefix, req.Path)
	}

	slog.Info("Scan requested", "path", req.Path, "full", containerPath)
	start := time.Now()
	ctx := r.Context()

	opts := orchestrator.Options{
		ProjectPath: containerPath,
		ForceStack:  req.Stack,
		RunExternal: req.External,
		ProbeHost:   "localhost",
	}

	rich := orchestrator.RunAllScanners(ctx, opts)
	s.lastResult = &rich

	// Initialize LLM if config is available
	if rich.Report.Config != nil && s.llmClient == nil {
		client, err := llm.NewClient(llm.Config{
			Provider: rich.Report.Config.LLM.Provider,
			Model:    rich.Report.Config.LLM.Model,
			APIKey:   rich.Report.Config.LLM.APIKey,
			BaseURL:  rich.Report.Config.LLM.BaseURL,
			Timeout:  rich.Report.Config.LLM.Timeout,
		})
		if err == nil {
			s.llmClient = client
		}
	}

	// ── Persist to DB ────────────────────────────────────────────────
	// 1. Auto-create or find product
	productID, err := s.productRepo.FindOrCreateByPath(ctx, req.Path)
	if err != nil {
		slog.Error("Failed to find/create product for scan", "path", req.Path, "error", err)
	}

	// 2. Create engagement
	scanPath := req.Path
	engagementName := fmt.Sprintf("Web Audit — %s", filepath.Base(req.Path))
	engagement := &models.Engagement{
		ProductID:      productID,
		Name:           engagementName,
		ScanPath:       &scanPath,
		EngagementType: "interactive",
		Status:         "in_progress",
	}
	if err := s.engagementRepo.Create(ctx, engagement); err != nil {
		slog.Error("Failed to create engagement", "error", err)
	}
	engagementID := engagement.ID

	// 3. Convert findings and bulk-insert — ALL scanner types
	var findings []findingDTO
	var dbFindings []models.Finding

	// 3a. Core SAST findings
	for _, result := range rich.Report.Results {
		findings = append(findings, findingDTO{
			ID:          result.ID,
			Name:        result.Name,
			Severity:    result.Severity,
			File:        result.File,
			Line:        result.Line,
			Suggestion:  result.Suggestion,
			OWASP:       result.OWASPMapping,
			AuditStatus: string(result.AuditStatus),
			Stack:       result.Framework,
		})

		filePath := result.File
		lineNum := result.Line
		description := result.Suggestion
		dbFindings = append(dbFindings, models.Finding{
			EngagementID:  engagementID,
			ProductID:     &productID,
			RuleID:        result.ID,
			Title:         result.Name,
			Severity:      strings.ToUpper(result.Severity),
			FilePath:      &filePath,
			LineNumber:    &lineNum,
			Description:   &description,
			FixSuggestion: &result.Suggestion,
			Status:        "open",
			KanbanColumn:  "backlog",
			Stack:         result.Framework,
		})
	}

	// 3b. External scanner findings (Semgrep, Trivy, Bandit, Gitleaks)
	for _, ext := range rich.External {
		sev := strings.ToUpper(ext.Severity)
		if sev == "" {
			sev = "MEDIUM"
		}
		filePath := ext.File
		lineNum := ext.Line
		desc := ext.Message
		suggestion := ext.Message
		name := fmt.Sprintf("[%s] %s", ext.Source, ext.RuleID)
		if ext.RuleID == "" {
			name = fmt.Sprintf("[%s] %s", ext.Source, ext.Message)
		}

		findings = append(findings, findingDTO{
			ID:         ext.RuleID,
			Name:       name,
			Severity:   strings.ToLower(sev),
			File:       ext.File,
			Line:       ext.Line,
			Suggestion: ext.Message,
			Stack:      ext.Source,
		})
		dbFindings = append(dbFindings, models.Finding{
			EngagementID:  engagementID,
			ProductID:     &productID,
			RuleID:        ext.RuleID,
			Title:         name,
			Severity:      sev,
			FilePath:      &filePath,
			LineNumber:    &lineNum,
			Description:   &desc,
			FixSuggestion: &suggestion,
			Status:        "open",
			KanbanColumn:  "backlog",
			Stack:         ext.Source,
		})
	}

	// 3c. NFR (Non-Functional Requirement) findings
	for _, n := range rich.NFR {
		sev := strings.ToUpper(n.Severity)
		if sev == "" {
			sev = "LOW"
		}
		desc := n.Message
		zeroLine := 0
		findings = append(findings, findingDTO{
			ID:         n.RuleID,
			Name:       n.Name,
			Severity:   strings.ToLower(sev),
			Suggestion: n.Message,
			Stack:      "nfr",
		})
		dbFindings = append(dbFindings, models.Finding{
			EngagementID:  engagementID,
			ProductID:     &productID,
			RuleID:        n.RuleID,
			Title:         n.Name,
			Severity:      sev,
			LineNumber:    &zeroLine,
			Description:   &desc,
			FixSuggestion: &desc,
			Status:        "open",
			KanbanColumn:  "backlog",
			Stack:         "nfr",
		})
	}

	// 3d. Deploy / IaC findings
	for _, d := range rich.Deploy {
		sev := strings.ToUpper(d.Severity)
		if sev == "" {
			sev = "MEDIUM"
		}
		filePath := d.File
		lineNum := d.Line
		desc := d.Issue
		advice := d.Advice
		findings = append(findings, findingDTO{
			ID:         fmt.Sprintf("deploy-%s-%d", filepath.Base(d.File), d.Line),
			Name:       d.Issue,
			Severity:   strings.ToLower(sev),
			File:       d.File,
			Line:       d.Line,
			Suggestion: d.Advice,
			Stack:      "deploy",
		})
		dbFindings = append(dbFindings, models.Finding{
			EngagementID:  engagementID,
			ProductID:     &productID,
			RuleID:        fmt.Sprintf("deploy-%s-%d", filepath.Base(d.File), d.Line),
			Title:         d.Issue,
			Severity:      sev,
			FilePath:      &filePath,
			LineNumber:    &lineNum,
			Description:   &desc,
			FixSuggestion: &advice,
			Status:        "open",
			KanbanColumn:  "backlog",
			Stack:         "deploy",
		})
	}

	// 3e. Network probe findings
	for _, n := range rich.Network {
		sev := strings.ToUpper(n.Severity)
		if sev == "" {
			sev = "HIGH"
		}
		desc := n.Message
		host := n.Target
		zeroLine := 0
		name := fmt.Sprintf("Open port %d (%s) on %s", n.Port, n.Service, n.Target)
		findings = append(findings, findingDTO{
			ID:         fmt.Sprintf("net-%s-%d", n.Target, n.Port),
			Name:       name,
			Severity:   strings.ToLower(sev),
			Suggestion: n.Message,
			Stack:      "network",
		})
		dbFindings = append(dbFindings, models.Finding{
			EngagementID:  engagementID,
			ProductID:     &productID,
			RuleID:        fmt.Sprintf("net-%s-%d", n.Target, n.Port),
			Title:         name,
			Severity:      sev,
			FilePath:      &host,
			LineNumber:    &zeroLine,
			Description:   &desc,
			FixSuggestion: &desc,
			Status:        "open",
			KanbanColumn:  "backlog",
			Stack:         "network",
		})
	}

	// 3f. Git history leaks
	for _, hl := range rich.HistoryLeaks {
		filePath := hl.FilePath
		zeroLine := 0
		desc := fmt.Sprintf("Pattern '%s' found in commit %s by %s: %s", hl.Pattern, hl.CommitHash, hl.Author, hl.LinePreview)
		findings = append(findings, findingDTO{
			ID:         fmt.Sprintf("gitleak-%s", hl.CommitHash[:8]),
			Name:       fmt.Sprintf("Secret leak: %s in %s", hl.Pattern, hl.FilePath),
			Severity:   "high",
			File:       hl.FilePath,
			Suggestion: desc,
			Stack:      "git-history",
		})
		dbFindings = append(dbFindings, models.Finding{
			EngagementID:  engagementID,
			ProductID:     &productID,
			RuleID:        fmt.Sprintf("gitleak-%s", hl.CommitHash[:8]),
			Title:         fmt.Sprintf("Secret leak: %s in %s", hl.Pattern, hl.FilePath),
			Severity:      "HIGH",
			FilePath:      &filePath,
			LineNumber:    &zeroLine,
			Description:   &desc,
			FixSuggestion: &desc,
			Status:        "open",
			KanbanColumn:  "backlog",
			Stack:         "git-history",
		})
	}

	if len(dbFindings) > 0 {
		if err := s.findingRepo.BulkCreate(ctx, dbFindings); err != nil {
			slog.Error("Failed to bulk-insert findings", "count", len(dbFindings), "error", err)
		} else {
			slog.Info("Findings persisted to database", "count", len(dbFindings), "engagement_id", engagementID)
		}
	}

	// ── Populate Topology ───────────────────────────────────────────
	_ = s.topologyRepo.Clear()

	appNodeID := fmt.Sprintf("app-%d", productID)
	appRisk := rich.Report.SecurityGrade
	if appRisk == "F" || appRisk == "D" {
		appRisk = "CRITICAL"
	} else if appRisk == "C" {
		appRisk = "MEDIUM"
	} else {
		appRisk = "LOW"
	}

	components := architect.DetectComponents(containerPath)
	mainAppName := filepath.Base(req.Path)

	for _, c := range components {
		if c.Type == "app" {
			mainAppName = fmt.Sprintf("%s (%s)", c.Name, filepath.Base(req.Path))
			break
		}
	}

	_ = s.topologyRepo.Upsert(repositories.TopologyNode{
		ID:     appNodeID,
		Name:   mainAppName,
		Type:   "APPLICATION",
		Status: "ONLINE",
		Risk:   appRisk,
	})

	for _, c := range components {
		if c.Type == "app" {
			continue
		}

		nodeType := "SYSTEM_NODE"
		switch c.Type {
		case "db":
			nodeType = "DATABASE"
		case "cache":
			nodeType = "CACHE"
		case "proxy":
			nodeType = "PROXY"
		case "storage":
			nodeType = "STORAGE"
		case "message_broker":
			nodeType = "MESSAGE_BROKER"
		}

		compRisk := "LOW"
		maxSeverityVal := 0
		severityMap := map[string]int{
			"LOW":      0,
			"INFO":     0,
			"MEDIUM":   1,
			"HIGH":     2,
			"CRITICAL": 3,
		}

		for _, f := range dbFindings {
			fTitle := strings.ToLower(f.Title)
			fDesc := ""
			if f.Description != nil {
				fDesc = strings.ToLower(*f.Description)
			}
			compNameLower := strings.ToLower(c.Name)
			if strings.Contains(fTitle, compNameLower) || strings.Contains(fDesc, compNameLower) {
				sevUpper := strings.ToUpper(f.Severity)
				val := severityMap[sevUpper]
				if val > maxSeverityVal {
					maxSeverityVal = val
					compRisk = sevUpper
				}
			}
		}

		compNodeID := fmt.Sprintf("infra-%d-%s", productID, strings.ToLower(strings.ReplaceAll(c.Name, " ", "-")))
		_ = s.topologyRepo.Upsert(repositories.TopologyNode{
			ID:     compNodeID,
			Name:   c.Name,
			Type:   nodeType,
			Status: "ONLINE",
			Risk:   compRisk,
		})
		_ = s.topologyRepo.UpsertLink(repositories.TopologyLink{Source: appNodeID, Target: compNodeID})
	}

	// 4. Mark engagement completed
	if err := s.engagementRepo.UpdateStatus(ctx, engagementID, "completed"); err != nil {
		slog.Error("Failed to update engagement status", "error", err)
	}

	var stacks []string
	for _, st := range rich.Report.Stacks {
		stacks = append(stacks, string(st))
	}

	duration := time.Since(start).String()
	slog.Info("Scan completed", "path", req.Path, "findings", len(findings), "duration", duration, "product_id", productID, "engagement_id", engagementID)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(scanResponse{
		Ok:            true,
		ScanID:        fmt.Sprintf("SCAN-%d", time.Now().Unix()),
		Findings:      findings,
		Dependencies:  rich.Report.Dependencies,
		Stacks:        stacks,
		SecurityScore: rich.Report.SecurityScore,
		SecurityGrade: rich.Report.SecurityGrade,
		HealthCheck:   rich.Report.HealthCheck,
		Duration:      duration,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	err := s.db.PingContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "database connection failed"})
		return
	}

	tools := map[string]bool{
		"semgrep":  external.IsInstalled("semgrep"),
		"bandit":   external.IsInstalled("bandit"),
		"gitleaks": external.IsInstalled("gitleaks"),
		"trivy":    external.IsInstalled("trivy"),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "tools": tools})
}

type browserEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Path  string `json:"path"`
}

func (s *Server) handleBrowser(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "."
	}

	fullPath := path
	if s.hostPrefix != "" && !strings.HasPrefix(path, s.hostPrefix) {
		fullPath = filepath.Join(s.hostPrefix, path)
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		slog.Error("Browser error", "path", fullPath, "err", err)
		if os.IsPermission(err) {
			jsonError(w, "Permission denied for this directory", http.StatusForbidden)
		} else if os.IsNotExist(err) {
			jsonError(w, "Directory not found", http.StatusNotFound)
		} else {
			jsonError(w, "Internal system error accessing directory", http.StatusInternalServerError)
		}
		return
	}

	var res []browserEntry
	for _, e := range entries {
		res = append(res, browserEntry{
			Name:  e.Name(),
			IsDir: e.IsDir(),
			Path:  filepath.Join(path, e.Name()),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"path":    path,
		"entries": res,
	})
}

// ── Admin Maintenance Endpoints ─────────────────────────────────────────────────────────────

func (s *Server) handleClearCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	slog.Warn("Admin requested to clear findings cache")
	// Clear scan results but keep topology nodes (or clear topology links to rescan).
	_, _ = s.db.Exec("DELETE FROM finding_notes")
	_, _ = s.db.Exec("DELETE FROM findings")
	_, _ = s.db.Exec("DELETE FROM engagements")
	_, _ = s.db.Exec("DELETE FROM topology_links")
	_, _ = s.db.Exec("DELETE FROM topology_nodes")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) handlePurgeDatabase(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	slog.Warn("Admin requested to PURGE ALL DATA")

	tables := []string{
		"finding_notes", "findings", "engagements", "product_members",
		"products", "product_types", "topology_links", "topology_nodes",
		"reports", "chat_messages", "chat_sessions", "audit_log", "notifications",
	}
	for _, table := range tables {
		_, err := s.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			slog.Error("Failed to purge table", "table", table, "error", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) handleRebuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	slog.Warn("Admin requested to REBUILD container - server will restart")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})

	go func() {
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
}

func (s *Server) handleTriage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Project string `json:"project"`
		ID      string `json:"id"`
		File    string `json:"file"`
		Action  string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid triage request", http.StatusBadRequest)
		return
	}

	fullProject := req.Project
	if s.hostPrefix != "" && !strings.HasPrefix(req.Project, s.hostPrefix) {
		fullProject = filepath.Join(s.hostPrefix, req.Project)
	}

	auditStore := core.NewAuditStore(fullProject)
	status := core.AuditStatusOpen
	if req.Action == "IGNORE" {
		status = core.AuditStatusIgnored
	} else if req.Action == "FIX" || req.Action == "TRIAGE" {
		status = core.AuditStatusTriage
	}

	auditStore.SetStatus(req.ID, req.File, status, "Triage via Web UI")
	err := auditStore.Save()
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if s.llmClient == nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": "AI Consultant is offline. Please provide a GEMINI_API_KEY.",
		})
		return
	}

	var req struct {
		Messages []llm.Message `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid chat request", http.StatusBadRequest)
		return
	}

	// Build security context from findings database
	ctx := r.Context()
	var systemCtx strings.Builder
	systemCtx.WriteString("You are AITriage Security Consultant — an expert application security engineer.\n")
	systemCtx.WriteString("You have access to the scan results for the user's repositories.\n\n")

	// Get product (project) list
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, COALESCE(repo_url,'') FROM products ORDER BY name`)
	if err == nil {
		defer rows.Close()
		systemCtx.WriteString("## Scanned Projects\n")
		for rows.Next() {
			var id int64
			var name, repoURL string
			_ = rows.Scan(&id, &name, &repoURL)
			// Count findings per severity for this product
			var crit, high, med, low int
			_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN severity='critical' THEN 1 ELSE 0 END),0), COALESCE(SUM(CASE WHEN severity='high' THEN 1 ELSE 0 END),0), COALESCE(SUM(CASE WHEN severity='medium' THEN 1 ELSE 0 END),0), COALESCE(SUM(CASE WHEN severity='low' THEN 1 ELSE 0 END),0) FROM findings WHERE product_id=?`, id).Scan(&crit, &high, &med, &low)
			systemCtx.WriteString(fmt.Sprintf("- **%s** (path: %s): %d critical, %d high, %d medium, %d low\n", name, repoURL, crit, high, med, low))
		}
		systemCtx.WriteString("\n")
	}

	// Get top critical/high findings (up to 30 for context)
	frows, err := s.db.QueryContext(ctx, `
		SELECT f.title, f.severity, COALESCE(f.file_path,''), COALESCE(f.line_number,0), COALESCE(f.description,''), COALESCE(p.name,'unknown')
		FROM findings f LEFT JOIN products p ON f.product_id = p.id
		WHERE f.severity IN ('critical','high') AND f.status NOT IN ('triage','false_positive','risk_accepted')
		ORDER BY CASE f.severity WHEN 'critical' THEN 0 WHEN 'high' THEN 1 END, f.id
		LIMIT 30
	`)
	if err == nil {
		defer frows.Close()
		systemCtx.WriteString("## Top Critical & High Findings\n")
		for frows.Next() {
			var title, severity, filePath, desc, project string
			var lineNum int
			_ = frows.Scan(&title, &severity, &filePath, &lineNum, &desc, &project)
			loc := filePath
			if lineNum > 0 {
				loc = fmt.Sprintf("%s:%d", filePath, lineNum)
			}
			systemCtx.WriteString(fmt.Sprintf("- [%s][%s] **%s** at `%s`\n", strings.ToUpper(severity), project, title, loc))
			if desc != "" && len(desc) < 200 {
				systemCtx.WriteString(fmt.Sprintf("  %s\n", desc))
			}
		}
		systemCtx.WriteString("\n")
	}

	// Get overall stats
	var totalFindings, openFindings int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM findings`).Scan(&totalFindings)
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM findings WHERE status NOT IN ('triage','false_positive','risk_accepted')`).Scan(&openFindings)
	systemCtx.WriteString(fmt.Sprintf("## Stats: %d total findings, %d open/active.\n\n", totalFindings, openFindings))
	systemCtx.WriteString("Answer in the user's language. Provide actionable remediation advice with code examples when possible. Reference specific files and line numbers from the findings.\n")

	// Prepend system message with context
	messages := make([]llm.Message, 0, len(req.Messages)+1)
	messages = append(messages, llm.Message{Role: "system", Content: systemCtx.String()})
	messages = append(messages, req.Messages...)

	reply, _, err := s.llmClient.Chat(ctx, messages)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"content": reply,
	})
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if s.llmClient == nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": "AI Consultant is offline.",
		})
		return
	}

	var req struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid analysis request", http.StatusBadRequest)
		return
	}

	var fTitle, fDesc, fCode, fFile, fSeverity string
	var fLine int
	findingID, err := strconv.ParseInt(req.ID, 10, 64)
	if err == nil {
		finding, err := s.findingRepo.GetByID(r.Context(), findingID)
		if err == nil && finding != nil {
			fTitle = finding.Title
			fSeverity = finding.Severity
			if finding.Description != nil {
				fDesc = *finding.Description
			}
			if finding.CodeSnippet != nil {
				fCode = *finding.CodeSnippet
			}
			if finding.FilePath != nil {
				fFile = *finding.FilePath
			}
			if finding.LineNumber != nil {
				fLine = *finding.LineNumber
			}
		}
	}

	prompt := fmt.Sprintf(`Please analyze this security finding:
Title: %s
Severity: %s
File: %s:%d
Description: %s
Code Snippet:
%s

Your task is to:
1. Determine if this finding is a True Positive (valid vulnerability) or a False Positive. Explain your reasoning in detail.
2. Provide a detailed, context-aware remediation plan if it is a True Positive, with clear fixed code examples.
3. Suggest a verification plan to check if the fix is correct.`, fTitle, fSeverity, fFile, fLine, fDesc, fCode)

	messages := []llm.Message{
		{
			Role:    "system",
			Content: "You are an elite DevSecOps engineer and AI security auditor. Analyze the provided finding and determine if it is a True Positive, False Positive, or Needs Human Review. Formulate your response as a clear, professional assessment focusing on exploitability, impact, and business risk.",
		},
		{Role: "user", Content: prompt},
	}

	analysis, _, err := s.llmClient.Chat(r.Context(), messages)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"analysis": analysis,
	})
}

func (s *Server) handleAITriage(w http.ResponseWriter, r *http.Request) {
	if s.llmClient == nil {
		jsonError(w, "AI Consultant is offline. Please provide a GEMINI_API_KEY.", http.StatusServiceUnavailable)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/findings/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		jsonError(w, "missing finding id", http.StatusBadRequest)
		return
	}
	findingID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		jsonError(w, "invalid finding id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	finding, err := s.findingRepo.GetByID(ctx, findingID)
	if err != nil {
		jsonError(w, fmt.Sprintf("failed to get finding: %v", err), http.StatusNotFound)
		return
	}

	slog.Info("AI Triage started", "finding_id", findingID, "title", finding.Title, "severity", finding.Severity)

	// 1. Resolve code file context
	var fileContent string
	var fullPath string
	if finding.FilePath != nil && *finding.FilePath != "" {
		eng, err := s.engagementRepo.GetByID(ctx, finding.EngagementID)
		if err == nil && eng != nil && eng.ScanPath != nil && *eng.ScanPath != "" {
			fullPath = filepath.Join(*eng.ScanPath, *finding.FilePath)
		} else {
			prod, err := s.productRepo.GetByID(ctx, *finding.ProductID)
			if err == nil && prod != nil && prod.RepoURL != nil && *prod.RepoURL != "" {
				fullPath = filepath.Join(*prod.RepoURL, *finding.FilePath)
			} else {
				fullPath = *finding.FilePath
			}
		}
		if s.hostPrefix != "" && !strings.HasPrefix(fullPath, s.hostPrefix) {
			fullPath = filepath.Join(s.hostPrefix, fullPath)
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			slog.Warn("AI Triage: could not read source file", "path", fullPath, "error", err)
		} else {
			lines := strings.Split(string(data), "\n")
			start := 0
			end := len(lines)
			if finding.LineNumber != nil && *finding.LineNumber > 0 {
				lineIdx := *finding.LineNumber - 1
				start = lineIdx - 40
				if start < 0 {
					start = 0
				}
				end = lineIdx + 40
				if end > len(lines) {
					end = len(lines)
				}
				var snippet []string
				for i := start; i < end; i++ {
					prefix := "   "
					if i == lineIdx {
						prefix = ">> "
					}
					snippet = append(snippet, fmt.Sprintf("%s%d: %s", prefix, i+1, lines[i]))
				}
				fileContent = strings.Join(snippet, "\n")
			} else {
				// Limit to first 200 lines if no specific line
				if len(lines) > 200 {
					fileContent = strings.Join(lines[:200], "\n") + "\n... (truncated)"
				} else {
					fileContent = string(data)
				}
			}
		}
	}

	// 2. Collect ALL available finding metadata for rich context
	fTitle := finding.Title
	fSeverity := finding.Severity
	fRuleID := finding.RuleID
	fStack := finding.Stack

	// Fetch project (product) and engagement for full context
	var productName, productTechStack, productPlatform, productCriticality string
	var engagementName, scanPath string
	if finding.ProductID != nil {
		prod, err := s.productRepo.GetByID(ctx, *finding.ProductID)
		if err == nil && prod != nil {
			productName = prod.Name
			if prod.TechStack != nil {
				productTechStack = *prod.TechStack
			}
			if prod.Platform != nil {
				productPlatform = *prod.Platform
			}
			productCriticality = prod.BusinessCriticality
		}
	}
	eng, err := s.engagementRepo.GetByID(ctx, finding.EngagementID)
	if err == nil && eng != nil {
		engagementName = eng.Name
		if eng.ScanPath != nil {
			scanPath = *eng.ScanPath
		}
	}

	// Build a comprehensive context block
	fFile := ""
	if finding.FilePath != nil {
		fFile = *finding.FilePath
	}
	fLine := 0
	if finding.LineNumber != nil {
		fLine = *finding.LineNumber
	}
	fDesc := ""
	if finding.Description != nil {
		fDesc = *finding.Description
	}
	fCode := ""
	if finding.CodeSnippet != nil {
		fCode = *finding.CodeSnippet
	}
	fFixSuggestion := ""
	if finding.FixSuggestion != nil {
		fFixSuggestion = *finding.FixSuggestion
	}
	fImpact := ""
	if finding.Impact != nil {
		fImpact = *finding.Impact
	}
	fCWE := ""
	if finding.CWEID != nil {
		fCWE = *finding.CWEID
	}
	fCVE := ""
	if finding.CVEID != nil {
		fCVE = *finding.CVEID
	}

	var contextParts []string
	if productName != "" {
		contextParts = append(contextParts, fmt.Sprintf("Project: %s", productName))
	}
	if productTechStack != "" {
		contextParts = append(contextParts, fmt.Sprintf("Tech Stack: %s", productTechStack))
	}
	if productPlatform != "" {
		contextParts = append(contextParts, fmt.Sprintf("Platform: %s", productPlatform))
	}
	if productCriticality != "" {
		contextParts = append(contextParts, fmt.Sprintf("Business Criticality: %s", productCriticality))
	}
	if engagementName != "" {
		contextParts = append(contextParts, fmt.Sprintf("Engagement: %s", engagementName))
	}
	if scanPath != "" {
		contextParts = append(contextParts, fmt.Sprintf("Scan Path: %s", scanPath))
	}
	contextParts = append(contextParts, fmt.Sprintf("Title: %s", fTitle))
	contextParts = append(contextParts, fmt.Sprintf("Severity: %s", fSeverity))
	contextParts = append(contextParts, fmt.Sprintf("Scanner / Stack: %s", fStack))
	contextParts = append(contextParts, fmt.Sprintf("Rule ID: %s", fRuleID))
	if fFile != "" {
		contextParts = append(contextParts, fmt.Sprintf("File Path: %s", fFile))
	}
	if fLine > 0 {
		contextParts = append(contextParts, fmt.Sprintf("Line Number: %d", fLine))
	}
	if fCWE != "" {
		contextParts = append(contextParts, fmt.Sprintf("CWE: %s", fCWE))
	}
	if fCVE != "" {
		contextParts = append(contextParts, fmt.Sprintf("CVE: %s", fCVE))
	}
	if fDesc != "" {
		contextParts = append(contextParts, fmt.Sprintf("Description: %s", fDesc))
	}
	if fFixSuggestion != "" {
		contextParts = append(contextParts, fmt.Sprintf("Remediation / Fix Suggestion: %s", fFixSuggestion))
	}
	if fImpact != "" {
		contextParts = append(contextParts, fmt.Sprintf("Impact: %s", fImpact))
	}
	if fCode != "" {
		contextParts = append(contextParts, fmt.Sprintf("Code Snippet from Scanner:\n```\n%s\n```", fCode))
	}

	findingContext := strings.Join(contextParts, "\n")

	var codeContextBlock string
	if fileContent != "" {
		codeContextBlock = fmt.Sprintf("Source Code Context (from %s around line %d):\n```\n%s\n```", fFile, fLine, fileContent)
	} else {
		codeContextBlock = "Source Code Context: Not available (file could not be read). Perform triage based on the finding details, vulnerability type, project context, and your security expertise."
	}

	prompt := fmt.Sprintf(`## Vulnerability Finding
%s

## %s

## Triage Methodology

Evaluate this finding using the following criteria (based on SecureCoder threat model analysis):

### 1. Reachability Analysis
- Is the flagged code reachable from an **untrusted entry point** (HTTP handler, CLI arg, file input, user-controlled data)?
- Or is it only reachable from trusted internal code paths?

### 2. Trust Boundary & Auth Context
- Does the project's **authentication/authorization** context mitigate the risk?
- Are there **implicit trust assumptions** (e.g., "only internal services call this") that make exploitation unlikely?
- Does the data cross **trust boundaries** (frontend→backend, user→admin)?

### 3. Exploitability Assessment
- Is the vulnerability **actually exploitable** given the deployment context (web app, CLI tool, internal service)?
- Are there existing **mitigations** in place (input validation, sanitization, parameterized queries, CSP headers, rate limiting)?
- Would an attacker realistically be able to trigger this code path with malicious input?

### 4. Vulnerability-Specific Checks
For common vulnerability types, check:
- **Path Traversal**: Does the code normalize or reject "../" sequences? Is the resolved path validated?
- **XSS**: Is user input escaped/sanitized before DOM insertion? Are Content-Security-Policy headers set?
- **SQL Injection**: Are parameterized queries/prepared statements used? Or is there string concatenation?
- **SSRF**: Are target URLs validated and restricted? Are internal IP ranges blocked?
- **Hardcoded Secrets**: Is this a real secret or a placeholder/test value? Is it in a test/example file?
- **Missing Rate Limiting**: Is this an authentication endpoint or public-facing API that needs rate limiting?
- **Insecure Deserialization**: Are safe deserialization methods used? Are types restricted?

### 5. Classification

| Disposition | Criteria |
|---|---|
| **True Positive** | Code IS reachable from untrusted input, vulnerability IS exploitable, no sufficient mitigations exist |
| **False Positive** | Code is NOT reachable from untrusted input, OR mitigations already exist, OR scanner pattern match is incorrect, OR this is intended functionality |
| **Needs Review** | Insufficient context to determine reachability or exploitability; requires manual security engineer review |

## Response Format
Return ONLY a valid JSON object with no other text:
{
  "status": "true_positive" | "false_positive" | "needs_review",
  "summary": "Конкретное объяснение на русском (1-3 предложения): укажите ПОЧЕМУ — какой именно код уязвим/защищён, какие entry points задействованы, какие митигации есть или отсутствуют"
}`, findingContext, codeContextBlock)

	slog.Info("AI Triage prompt built", "finding_id", findingID, "has_code_context", fileContent != "", "prompt_len", len(prompt))

	messages := []llm.Message{
		{
			Role:    "system",
			Content: "You are an elite application security engineer performing automated vulnerability triage. You use the SecureCoder threat model methodology: you analyze entry points, trust boundaries, auth context, and exploitability to classify scanner findings. You are precise and reference specific code patterns. You output ONLY valid JSON.",
		},
		{Role: "user", Content: prompt},
	}

	reply, _, err := s.llmClient.Chat(ctx, messages)
	if err != nil {
		slog.Error("AI Triage LLM error", "finding_id", findingID, "error", err)
		jsonError(w, fmt.Sprintf("LLM chat error: %v", err), http.StatusInternalServerError)
		return
	}

	slog.Info("AI Triage LLM response received", "finding_id", findingID, "reply_len", len(reply))

	// Parse JSON response — handle ```json blocks and bare JSON
	cleaned := reply
	if idx := strings.Index(cleaned, "```json"); idx != -1 {
		cleaned = cleaned[idx+7:]
		if endIdx := strings.Index(cleaned, "```"); endIdx != -1 {
			cleaned = cleaned[:endIdx]
		}
	} else if idx := strings.Index(cleaned, "```"); idx != -1 {
		cleaned = cleaned[idx+3:]
		if endIdx := strings.Index(cleaned, "```"); endIdx != -1 {
			cleaned = cleaned[:endIdx]
		}
	}
	// Also try to extract JSON from { to }
	cleaned = strings.TrimSpace(cleaned)
	if !strings.HasPrefix(cleaned, "{") {
		if braceIdx := strings.Index(cleaned, "{"); braceIdx != -1 {
			cleaned = cleaned[braceIdx:]
		}
	}
	if lastBrace := strings.LastIndex(cleaned, "}"); lastBrace != -1 {
		cleaned = cleaned[:lastBrace+1]
	}

	var triageRes struct {
		Status  string `json:"status"`
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(cleaned), &triageRes); err != nil {
		slog.Warn("AI Triage JSON parse failed, using fallback", "finding_id", findingID, "error", err, "raw", reply[:min(len(reply), 200)])
		triageRes.Status = "needs_review"
		// Extract something useful from the raw reply
		triageRes.Summary = "AI анализ не смог вернуть структурированный ответ. Необходима ручная проверка. Ответ AI: " + reply
		if len(triageRes.Summary) > 500 {
			triageRes.Summary = triageRes.Summary[:500] + "..."
		}
	}

	// Validate status
	switch triageRes.Status {
	case "true_positive", "false_positive", "needs_review":
		// valid
	default:
		slog.Warn("AI Triage returned unknown status, defaulting to needs_review", "finding_id", findingID, "status", triageRes.Status)
		triageRes.Status = "needs_review"
	}

	slog.Info("AI Triage completed", "finding_id", findingID, "status", triageRes.Status)

	// 3. Update database
	err = s.findingRepo.UpdateAITriage(ctx, findingID, triageRes.Status, triageRes.Summary)
	if err != nil {
		jsonError(w, fmt.Sprintf("failed to update finding AI triage: %v", err), http.StatusInternalServerError)
		return
	}

	// If False Positive, automatically update the finding status
	if triageRes.Status == "false_positive" {
		_ = s.findingRepo.UpdateStatus(ctx, findingID, "false_positive")
		_, _ = s.db.ExecContext(ctx, "UPDATE findings SET is_false_positive = 1, fp_reason = ? WHERE id = ?", triageRes.Summary, findingID)
	} else if triageRes.Status == "true_positive" {
		_ = s.findingRepo.UpdateStatus(ctx, findingID, "triage")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"status":  triageRes.Status,
		"summary": triageRes.Summary,
	})
}

func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rules := make([]engine.Rule, 0)
	if s.engine != nil {
		rules = s.engine.Rules
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":    true,
		"rules": rules,
	})
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		jsonError(w, "path is required", http.StatusBadRequest)
		return
	}

	fullPath := path
	if s.hostPrefix != "" && !strings.HasPrefix(path, s.hostPrefix) {
		fullPath = filepath.Join(s.hostPrefix, path)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"content": string(content),
	})
}

func (s *Server) Listen(addr string) error {
	slog.Info("AITriage Web UI started", "url", "http://"+addr)
	return http.ListenAndServe(addr, s)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	productIDStr := r.URL.Query().Get("product_id")
	var productID int
	var errConv error
	useProduct := false
	if productIDStr != "" {
		productID, errConv = strconv.Atoi(productIDStr)
		if errConv == nil {
			useProduct = true
		}
	}

	var openCount, critCount, highCount, medCount, lowCount int
	var findingTitles []string
	var err error

	if useProduct {
		_ = s.db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM findings 
			WHERE status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive') AND product_id = ?
		`, productID).Scan(&openCount)

		_ = s.db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM findings 
			WHERE status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive') AND severity = 'CRITICAL' AND product_id = ?
		`, productID).Scan(&critCount)

		_ = s.db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM findings 
			WHERE status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive') AND severity = 'HIGH' AND product_id = ?
		`, productID).Scan(&highCount)

		_ = s.db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM findings 
			WHERE status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive') AND severity = 'MEDIUM' AND product_id = ?
		`, productID).Scan(&medCount)

		_ = s.db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM findings 
			WHERE status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive') AND severity = 'LOW' AND product_id = ?
		`, productID).Scan(&lowCount)

		rows, errQuery := s.db.QueryContext(r.Context(), `
			SELECT title FROM findings 
			WHERE status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive') AND product_id = ?
			LIMIT 5
		`, productID)
		err = errQuery
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var t string
				if err := rows.Scan(&t); err == nil {
					findingTitles = append(findingTitles, t)
				}
			}
		}
	} else {
		_ = s.db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM findings 
			WHERE status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive')
		`).Scan(&openCount)

		_ = s.db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM findings 
			WHERE status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive') AND severity = 'CRITICAL'
		`).Scan(&critCount)

		_ = s.db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM findings 
			WHERE status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive') AND severity = 'HIGH'
		`).Scan(&highCount)

		_ = s.db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM findings 
			WHERE status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive') AND severity = 'MEDIUM'
		`).Scan(&medCount)

		_ = s.db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM findings 
			WHERE status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive') AND severity = 'LOW'
		`).Scan(&lowCount)

		rows, errQuery := s.db.QueryContext(r.Context(), `
			SELECT title FROM findings 
			WHERE status NOT IN ('resolved', 'closed', 'risk_accepted', 'false_positive')
			LIMIT 5
		`)
		err = errQuery
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var t string
				if err := rows.Scan(&t); err == nil {
					findingTitles = append(findingTitles, t)
				}
			}
		}
	}

	var summary string
	if s.llmClient != nil && openCount > 0 {
		prompt := fmt.Sprintf("Project has %d open findings: %d critical, %d high, %d medium, %d low. Top finding titles: %s.", openCount, critCount, highCount, medCount, lowCount, strings.Join(findingTitles, ", "))
		messages := []llm.Message{
			{
				Role:    "system",
				Content: "You are an expert AI security engineer and consultant. Write a brief (max 3 sentences) security posture summary of the project in Russian. Do NOT use markdown. Start with a clear statement on the overall risk level.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		}
		reply, _, err := s.llmClient.Chat(r.Context(), messages)
		if err == nil && reply != "" {
			summary = reply
		}
	}

	if summary == "" {
		if openCount == 0 {
			summary = "В проекте не обнаружено активных уязвимостей. Система находится в номинальном состоянии."
		} else {
			summary = fmt.Sprintf("Анализ проекта завершен. Обнаружено %d активных уязвимостей (Critical: %d, High: %d, Medium: %d, Low: %d). Рекомендуется в первую очередь исправить критические уязвимости.", openCount, critCount, highCount, medCount, lowCount)
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"summary": summary,
	})
}

func (s *Server) handleChatSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		// List all sessions (user_id=1 for now since auth may be disabled)
		sessions, err := s.chatRepo.ListSessions(r.Context(), 1)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if sessions == nil {
			sessions = []repositories.ChatSession{}
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "sessions": sessions})
	case "POST":
		var req struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			req.Title = "New Chat"
		}
		if req.Title == "" {
			req.Title = "New Chat"
		}
		id, err := s.chatRepo.CreateSession(r.Context(), 1, req.Title)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "id": id})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleChatSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Extract session ID from URL: /api/chat/sessions/123
	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		jsonError(w, "missing session id", http.StatusBadRequest)
		return
	}
	idStr := parts[len(parts)-1]
	var sessionID int
	if _, err := fmt.Sscanf(idStr, "%d", &sessionID); err != nil {
		jsonError(w, "invalid session id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "DELETE":
		if err := s.chatRepo.DeleteSession(r.Context(), sessionID); err != nil {
			json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	case "PUT":
		var req struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
			jsonError(w, "title is required", http.StatusBadRequest)
			return
		}
		if err := s.chatRepo.UpdateSessionTitle(r.Context(), sessionID, req.Title); err != nil {
			json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleChatMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		sessionIDStr := r.URL.Query().Get("session_id")
		var sessionID int
		if _, err := fmt.Sscanf(sessionIDStr, "%d", &sessionID); err != nil {
			jsonError(w, "session_id is required", http.StatusBadRequest)
			return
		}
		msgs, err := s.chatRepo.GetMessages(r.Context(), sessionID)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if msgs == nil {
			msgs = []repositories.ChatMessage{}
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "messages": msgs})
	case "POST":
		var req struct {
			SessionID int    `json:"session_id"`
			Role      string `json:"role"`
			Content   string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request", http.StatusBadRequest)
			return
		}
		id, err := s.chatRepo.AddMessage(r.Context(), req.SessionID, req.Role, req.Content)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "id": id})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRunway(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case "GET":
		pidStr := r.URL.Query().Get("product_id")
		if pidStr == "" {
			// List all active sessions
			// For now just return empty
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "session": nil})
			return
		}
		pid, err := strconv.ParseInt(pidStr, 10, 64)
		if err != nil {
			jsonError(w, "invalid product_id", http.StatusBadRequest)
			return
		}
		session, err := s.runwayRepo.GetActiveByProductID(ctx, pid)
		if err != nil {
			// No active session
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "session": nil})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "session": session})

	case "POST":
		var req struct {
			ProductID int64 `json:"product_id"`
			AutoMode  bool  `json:"auto_mode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request", http.StatusBadRequest)
			return
		}
		if req.ProductID == 0 {
			jsonError(w, "product_id is required", http.StatusBadRequest)
			return
		}
		session := &models.RunwaySession{
			ProductID: req.ProductID,
			Status:    "in_progress",
			AutoMode:  req.AutoMode,
		}
		if err := s.runwayRepo.Create(ctx, session); err != nil {
			jsonError(w, fmt.Sprintf("failed to create session: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "session": session})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRunwaySession(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/runway/")
	if idStr == "" {
		jsonError(w, "missing session id", http.StatusBadRequest)
		return
	}

	// Handle /api/runway/history?product_id=X
	if idStr == "history" {
		pidStr := r.URL.Query().Get("product_id")
		if pidStr == "" {
			jsonError(w, "product_id is required", http.StatusBadRequest)
			return
		}
		pid, err := strconv.ParseInt(pidStr, 10, 64)
		if err != nil {
			jsonError(w, "invalid product_id", http.StatusBadRequest)
			return
		}
		sessions, err := s.runwayRepo.ListByProductID(r.Context(), pid)
		if err != nil {
			jsonError(w, fmt.Sprintf("failed to list sessions: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "sessions": sessions})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid session id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	switch r.Method {
	case "GET":
		session, err := s.runwayRepo.GetByID(ctx, id)
		if err != nil {
			jsonError(w, fmt.Sprintf("session not found: %v", err), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "session": session})

	case "PUT":
		session, err := s.runwayRepo.GetByID(ctx, id)
		if err != nil {
			jsonError(w, fmt.Sprintf("session not found: %v", err), http.StatusNotFound)
			return
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if val, ok := body["product_id"]; ok {
			if fval, ok := val.(float64); ok {
				session.ProductID = int64(fval)
			}
		}
		if val, ok := body["status"]; ok {
			if sval, ok := val.(string); ok {
				session.Status = sval
			}
		}
		if val, ok := body["current_step"]; ok {
			if fval, ok := val.(float64); ok {
				session.CurrentStep = int(fval)
			}
		}
		if val, ok := body["auto_mode"]; ok {
			if bval, ok := val.(bool); ok {
				session.AutoMode = bval
			}
		}
		if val, ok := body["threat_model"]; ok {
			if sval, ok := val.(string); ok {
				session.ThreatModel = &sval
			} else if val == nil {
				session.ThreatModel = nil
			}
		}
		if val, ok := body["security_plan"]; ok {
			if sval, ok := val.(string); ok {
				session.SecurityPlan = &sval
			} else if val == nil {
				session.SecurityPlan = nil
			}
		}
		if val, ok := body["remediation"]; ok {
			if sval, ok := val.(string); ok {
				session.Remediation = &sval
			} else if val == nil {
				session.Remediation = nil
			}
		}
		if val, ok := body["poc"]; ok {
			if sval, ok := val.(string); ok {
				session.PoC = &sval
			} else if val == nil {
				session.PoC = nil
			}
		}
		if val, ok := body["audit_report"]; ok {
			if sval, ok := val.(string); ok {
				session.AuditReport = &sval
			} else if val == nil {
				session.AuditReport = nil
			}
		}
		if val, ok := body["scan_count_before"]; ok {
			if fval, ok := val.(float64); ok {
				session.ScanCountBefore = int(fval)
			}
		}
		if val, ok := body["scan_count_after"]; ok {
			if fval, ok := val.(float64); ok {
				session.ScanCountAfter = int(fval)
			}
		}
		if val, ok := body["error_message"]; ok {
			if sval, ok := val.(string); ok {
				session.ErrorMessage = &sval
			} else if val == nil {
				session.ErrorMessage = nil
			}
		}

		if err := s.runwayRepo.Update(ctx, session); err != nil {
			jsonError(w, fmt.Sprintf("failed to update session: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "session": session})

	case "DELETE":
		if err := s.runwayRepo.Delete(ctx, id); err != nil {
			jsonError(w, fmt.Sprintf("failed to delete session: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRunwayAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	sessions, err := s.runwayRepo.ListAll(r.Context())
	if err != nil {
		jsonError(w, fmt.Sprintf("failed to list sessions: %v", err), http.StatusInternalServerError)
		return
	}

	// Enrich with product names
	type enrichedSession struct {
		models.RunwaySession
		ProductName string `json:"product_name"`
	}
	var result []enrichedSession
	for _, sess := range sessions {
		name := "Unknown"
		prod, err := s.productRepo.GetByID(r.Context(), sess.ProductID)
		if err == nil && prod != nil {
			name = prod.Name
		}
		result = append(result, enrichedSession{RunwaySession: sess, ProductName: name})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "sessions": result})
}

func (s *Server) handleRunwayExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/runway/export/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid session id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	session, err := s.runwayRepo.GetByID(ctx, id)
	if err != nil {
		jsonError(w, fmt.Sprintf("session not found: %v", err), http.StatusNotFound)
		return
	}

	product, err := s.productRepo.GetByID(ctx, session.ProductID)
	if err != nil || product == nil {
		jsonError(w, "product not found", http.StatusNotFound)
		return
	}

	// Build markdown content
	var md strings.Builder
	md.WriteString("# 🛡️ AITriage Security Audit Report\n\n")
	md.WriteString(fmt.Sprintf("**Project**: %s\n", product.Name))
	md.WriteString(fmt.Sprintf("**Date**: %s\n", session.CreatedAt.Format("2006-01-02 15:04")))
	md.WriteString(fmt.Sprintf("**Session ID**: %d\n", session.ID))
	md.WriteString(fmt.Sprintf("**Status**: %s\n", session.Status))
	md.WriteString(fmt.Sprintf("**Findings**: %d before → %d after\n\n", session.ScanCountBefore, session.ScanCountAfter))
	md.WriteString("---\n\n")

	if session.ThreatModel != nil && *session.ThreatModel != "" {
		md.WriteString("## 1. STRIDE Threat Model\n\n")
		md.WriteString(*session.ThreatModel)
		md.WriteString("\n\n---\n\n")
	}
	if session.SecurityPlan != nil && *session.SecurityPlan != "" {
		md.WriteString("## 2. Security Implementation Plan\n\n")
		md.WriteString(*session.SecurityPlan)
		md.WriteString("\n\n---\n\n")
	}
	if session.Remediation != nil && *session.Remediation != "" {
		md.WriteString("## 3. Remediation Patches\n\n")
		md.WriteString(*session.Remediation)
		md.WriteString("\n\n---\n\n")
	}
	if session.PoC != nil && *session.PoC != "" {
		md.WriteString("## 4. Proof of Concept Verification\n\n")
		md.WriteString(*session.PoC)
		md.WriteString("\n\n---\n\n")
	}
	if session.AuditReport != nil && *session.AuditReport != "" {
		md.WriteString("## 5. Audit Report\n\n")
		md.WriteString(*session.AuditReport)
		md.WriteString("\n\n---\n\n")
	}
	md.WriteString("\n*Generated by AITriage SecureCoder Agent*\n")

	mdContent := md.String()

	// Resolve project path and save file
	var projectPath string
	if product.RepoURL != nil && *product.RepoURL != "" {
		projectPath = *product.RepoURL
		if s.hostPrefix != "" && !strings.HasPrefix(projectPath, s.hostPrefix) {
			projectPath = filepath.Join(s.hostPrefix, projectPath)
		}
	}

	var savedPath string
	if projectPath != "" {
		aitriageDir := filepath.Join(projectPath, "aitriage")
		if err := os.MkdirAll(aitriageDir, 0755); err != nil {
			slog.Error("Failed to create aitriage directory", "path", aitriageDir, "error", err)
		} else {
			filename := fmt.Sprintf("runway-report-%d-%s.md", session.ID, session.CreatedAt.Format("2006-01-02"))
			fullPath := filepath.Join(aitriageDir, filename)
			if err := os.WriteFile(fullPath, []byte(mdContent), 0644); err != nil {
				slog.Error("Failed to write runway report", "path", fullPath, "error", err)
			} else {
				savedPath = filepath.Join("aitriage", filename)
				slog.Info("Runway report saved", "path", fullPath)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"content":  mdContent,
		"saved_to": savedPath,
	})
}

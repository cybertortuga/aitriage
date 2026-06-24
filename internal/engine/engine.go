package engine

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/cybertortuga/aitriage/internal/config"
	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/engine/loader"
	"github.com/cybertortuga/aitriage/internal/models"
	"github.com/cybertortuga/aitriage/internal/scanner/ast"
	"github.com/cybertortuga/aitriage/rules"
	"gopkg.in/yaml.v3"
)

// Re-export Rule for convenience or just use models.Rule
type Rule = models.Rule

// ... (Rest of Engine and Ruleset)

func (e *Engine) LoadExternalRules(path string) error {
	newRules, err := loader.LoadSemgrepRules(path)
	if err != nil {
		return err
	}

	for _, r := range newRules {
		e.compileRule(&r)
		e.Rules = append(e.Rules, r)
	}
	return nil
}

func (e *Engine) compileRule(r *models.Rule) {
	if r.Pattern != "" {
		c, err := regexp.Compile(r.Pattern)
		if err == nil {
			r.CompiledPattern = c
		}
	}
	for i := range r.Patterns {
		e.compileRule(&r.Patterns[i])
	}
	for i := range r.PatternEither {
		e.compileRule(&r.PatternEither[i])
	}
}

// loadEmbeddedRules walks the rules.FS embedded filesystem and parses
// every .yaml file it finds, merging all rule sets into a single slice.
// This makes rules/ the single source of truth — no duplication needed.
func loadEmbeddedRules() ([]models.Rule, error) {
	var allRules []models.Rule
	seen := make(map[string]bool) // deduplicate by rule ID

	err := fs.WalkDir(rules.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		slog.Debug("Walking rules FS", "path", path, "isDir", d.IsDir())
		// Skip non-YAML files (README.md, embed.go, etc.)
		if d.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}

		data, err := fs.ReadFile(rules.FS, path)
		if err != nil {
			slog.Warn("Failed to read rule file", "path", path, "error", err)
			return nil // skip broken files, don't crash
		}

		var rs models.Ruleset
		if err := yaml.Unmarshal(data, &rs); err != nil {
			slog.Warn("Failed to parse rule file", "path", path, "error", err)
			return nil
		}

		for _, r := range rs.Rules {
			// Skip rules that are documentation-only (entropy-analysis is handled in Go code)
			if r.Target == "entropy-analysis" {
				continue
			}
			if !seen[r.ID] {
				seen[r.ID] = true
				allRules = append(allRules, r)
			}
		}

		return nil
	})

	slog.Debug("Embedded rules loaded", "count", len(allRules))
	return allRules, err
}

// loadPackRules loads rule YAML files from installed rule packs in ~/.aitriage/packs/.
func loadPackRules() ([]models.Rule, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil // silently skip if no home dir
	}
	packsDir := filepath.Join(home, ".aitriage", "packs")
	if _, err := os.Stat(packsDir); os.IsNotExist(err) {
		return nil, nil // no packs installed
	}

	var allRules []models.Rule
	seen := make(map[string]bool)

	err = filepath.WalkDir(packsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}
		// Skip manifest.json and non-rule files
		if d.Name() == "manifest.json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var rs models.Ruleset
		if err := yaml.Unmarshal(data, &rs); err != nil {
			slog.Warn("Failed to parse pack rule file", "path", path, "error", err)
			return nil
		}

		for _, r := range rs.Rules {
			if r.Target == "entropy-analysis" {
				continue
			}
			if !seen[r.ID] {
				seen[r.ID] = true
				allRules = append(allRules, r)
			}
		}
		return nil
	})

	return allRules, err
}

type Ruleset = models.Ruleset

type Engine struct {
	Rules  []models.Rule
	Config *config.Config
}

func NewEngine(cfg *config.Config) (*Engine, error) {
	loadedRules, err := loadEmbeddedRules()
	if err != nil {
		slog.Error("Failed to load embedded rules", "error", err)
		return nil, fmt.Errorf("failed to load embedded rules: %w", err)
	}

	e := &Engine{
		Config: cfg,
	}

	// Precompile patterns and parse severity
	for i, r := range loadedRules {
		// Parse Severity from Suggestion if not set
		if r.Severity == "" {
			parts := strings.SplitN(r.Suggestion, ": ", 2)
			if len(parts) == 2 {
				sel := strings.ToUpper(parts[0])
				if sel == "CRITICAL" || sel == "HIGH" || sel == "MEDIUM" || sel == "LOW" {
					loadedRules[i].Severity = sel
					loadedRules[i].Suggestion = parts[1]
				}
			}
		}

		if loadedRules[i].Severity == "" {
			loadedRules[i].Severity = "LOW" // Default fallback
		}

		e.compileRule(&loadedRules[i])
	}
	e.Rules = loadedRules

	// Load rule packs from ~/.aitriage/packs/
	if packRules, err := loadPackRules(); err == nil && len(packRules) > 0 {
		seen := make(map[string]bool, len(loadedRules))
		for _, r := range loadedRules {
			seen[r.ID] = true
		}
		for i, r := range packRules {
			if !seen[r.ID] {
				seen[r.ID] = true
				e.compileRule(&packRules[i])
				loadedRules = append(loadedRules, packRules[i])
			}
		}
		slog.Debug("Loaded rule packs", "count", len(packRules))
	}

	// Add Custom Rules from Config
	if cfg != nil {
		for _, cr := range cfg.CustomRules {
			r := Rule{
				ID:         cr.ID,
				Name:       cr.Name,
				Stack:      cr.Stack,
				Extensions: cr.Extensions,
				Target:     cr.Target,
				Pattern:    cr.Pattern,
				Condition:  cr.Condition,
				Files:      cr.Files,
				Suggestion: cr.Suggestion,
				Severity:   cr.Severity,
			}
			if r.Pattern != "" {
				c, err := regexp.Compile(r.Pattern)
				if err == nil {
					r.CompiledPattern = c
				}
			}
			loadedRules = append(loadedRules, r)
		}
	}

	return &Engine{Rules: loadedRules, Config: cfg}, nil
}

// Run executes the engine against a ProjectContext using a Single-Pass O(N) architecture.
func (e *Engine) Run(ctx *core.ProjectContext) []core.CheckResult {
	var results []core.CheckResult
	var mu sync.Mutex

	var projectRules []Rule
	var fileRules []Rule

	// 1. Filter and categorize rules
	for _, rule := range e.Rules {
		if ctx.Config != nil && ctx.Config.IsRuleIgnored(rule.ID) {
			continue
		}
		if rule.Stack != "universal" && rule.Stack != ctx.Stack {
			continue
		}

		if rule.Condition == "missing_lockfile" || rule.Condition == "missing" || rule.Condition == "required_pattern" {
			projectRules = append(projectRules, rule)
		} else {
			fileRules = append(fileRules, rule)
		}
	}

	// 2. Evaluate Project Rules (File metadata only)
	var wgProj sync.WaitGroup
	for _, rule := range projectRules {
		wgProj.Add(1)
		go func(r Rule) {
			defer wgProj.Done()
			e.evaluateProjectRule(r, ctx, &results, &mu)
		}(rule)
	}
	wgProj.Wait()

	// 3. Single-Pass Execution: Loop over N files and evaluate M rules + Heuristics
	var wgFile sync.WaitGroup
	for _, file := range ctx.Files {
		if file.IsBinary {
			continue
		}

		wgFile.Add(1)
		go func(f *core.FileInfo) {
			defer wgFile.Done()

			// A. Mapped File Rules
			for _, rule := range fileRules {
				if !e.matchesExtension(rule, f) {
					continue
				}

				evidence, line := e.evaluateFile(rule, f)
				if evidence != "" {
					mu.Lock()
					results = append(results, core.CheckResult{
						ID:         rule.ID,
						Name:       rule.Name,
						Status:     core.Absent,
						Evidence:   fmt.Sprintf("Rule %s found issue in %s", rule.ID, f.Path),
						Suggestion: rule.Suggestion,
						Framework:  ctx.Stack,
						Severity:   rule.Severity,
						Line:       line,
						File:       f.Path,
					})
					mu.Unlock()
				}
			}

			// B. Entropy Checks
			if !f.IsTest {
				if entropyRes := e.evaluateEntropy(f); len(entropyRes) > 0 {
					mu.Lock()
					results = append(results, entropyRes...)
					mu.Unlock()
				}
			}

			// C. Entropy Pattern Heuristics
			if entrRes := e.evaluateEntropyPattern(f); len(entrRes) > 0 {
				mu.Lock()
				results = append(results, entrRes...)
				mu.Unlock()
			}
		}(file)
	}

	wgFile.Wait()
	return results
}

func (e *Engine) matchesExtension(r Rule, f *core.FileInfo) bool {
	for _, ext := range r.Extensions {
		if strings.EqualFold(ext, f.Extension) || strings.EqualFold(ext, filepath.Base(f.Path)) {
			return true
		}
	}
	return false
}

// matchesExcludedPath returns true if the file path contains any of the rule's ExcludePaths fragments.
func matchesExcludedPath(r Rule, f *core.FileInfo) bool {
	for _, frag := range r.ExcludePaths {
		if strings.Contains(f.Path, frag) {
			return true
		}
	}
	return false
}

func (e *Engine) evaluateProjectRule(rule Rule, ctx *core.ProjectContext, results *[]core.CheckResult, mu *sync.Mutex) {
	if rule.Condition == "missing_lockfile" {
		hasLock := false

		// Check the filesystem directly at project root — reliable regardless of how ctx.Files is scoped
		for _, lockFile := range rule.Files {
			path := filepath.Join(ctx.RootPath, lockFile)
			if _, err := os.Stat(path); err == nil {
				hasLock = true
				break
			}
		}

		if !hasLock {
			mu.Lock()
			*results = append(*results, core.CheckResult{
				ID: rule.ID, Name: rule.Name, Status: core.Absent,
				Evidence:   "No lockfile found in " + ctx.RootPath,
				Suggestion: rule.Suggestion, Framework: rule.Stack, Severity: rule.Severity,
			})
			mu.Unlock()
		}
		return
	}

	if rule.Condition == "missing" && rule.Target == "filename" {
		found := false
		for _, ext := range rule.Extensions {
			files := ctx.FindFilesByExtension(ext)
			for _, f := range files {
				base := f.Path[strings.LastIndex(f.Path, "/")+1:]
				if rule.CompiledPattern != nil && rule.CompiledPattern.MatchString(base) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			mu.Lock()
			*results = append(*results, core.CheckResult{
				ID: rule.ID, Name: rule.Name, Status: core.Absent,
				Evidence:   "Missing required file matching: " + rule.Pattern,
				Suggestion: rule.Suggestion, Framework: rule.Stack, Severity: rule.Severity,
			})
			mu.Unlock()
		}
		return
	}

	if rule.Condition == "required_pattern" && rule.CompiledPattern != nil {
		found := false
		for _, ext := range rule.Extensions {
			files := ctx.FindFilesByExtension(ext)
			for _, f := range files {
				if rule.ExcludeTests && f.IsTest {
					continue
				}
				if matchesExcludedPath(rule, f) {
					continue
				}
				var b []byte
				switch rule.Target {
				case "code":
					b, _ = f.GetStrippedContent()
				case "docs":
					b, _ = f.GetDocsOnlyContent()
				default:
					b, _ = f.GetContent()
				}
				if rule.CompiledPattern.Match(b) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			mu.Lock()
			*results = append(*results, core.CheckResult{
				ID: rule.ID, Name: rule.Name, Status: core.Absent,
				Evidence:   "Required pattern not found: " + rule.Pattern,
				Suggestion: rule.Suggestion, Framework: rule.Stack, Severity: rule.Severity,
			})
			mu.Unlock()
		}
	}
}

// secretAssignRegex matches: const API_KEY = "xxx", secret: "xxx", token="xxx" etc.
// Only triggers when the variable name suggests a secret AND the value is a high-entropy string.
var placeholderRegex = regexp.MustCompile(`(?i)(todo|fixme|temporary|tempVar|result\d|var\d|placeholder)`)

var secretAssignRegex = regexp.MustCompile(
	`(?i)(?:^|[\s,({])(api[_-]?key|secret|token|password|passwd|private[_-]?key|auth[_-]?key|access[_-]?key|client[_-]?secret|signing[_-]?key)\s*[=:]\s*["']([a-zA-Z0-9+/=_\-]{32,})["']`,
)

// entropyAllowedExts — only scan actual code and .env files, never configs or docs.
var entropyAllowedExts = map[string]bool{
	".go": true, ".py": true,
	".ts": true, ".tsx": true, ".js": true, ".jsx": true,
	".env": true,
}

func (e *Engine) evaluateEntropy(f *core.FileInfo) []core.CheckResult {
	// Only scan code files — skip json, yaml, md, toml, etc.
	if !entropyAllowedExts[f.Extension] {
		return nil
	}
	// Skip test files — test fixtures legitimately contain dummy tokens
	if f.IsTest {
		return nil
	}

	content, err := f.GetContent()
	if err != nil {
		return nil
	}
	src := string(content)

	matches := secretAssignRegex.FindAllStringSubmatchIndex(src, -1)
	for _, loc := range matches {
		if len(loc) < 6 {
			continue
		}
		fullMatch := src[loc[0]:loc[1]]
		// Group 1 = variable name, Group 2 = value
		_ = fullMatch
		varName := src[loc[2]:loc[3]]
		val := src[loc[4]:loc[5]]

		// Must be genuinely high entropy — eliminates CSS classes, package names, URLs
		if !IsHighEntropy(val) {
			continue
		}

		// Skip strings that look like module paths or URLs (contain slashes)
		if strings.Contains(val, "/") {
			continue
		}

		// Skip obvious placeholder values
		lower := strings.ToLower(val)
		if strings.HasPrefix(lower, "your_") || strings.HasPrefix(lower, "change_me") ||
			strings.HasPrefix(lower, "replace_") || lower == "xxx" || lower == "secret" {
			continue
		}

		lineNumber := strings.Count(src[:loc[0]], "\n") + 1

		// Check inline suppression
		if e.isLineIgnored(f, lineNumber, "ENTROPY-SECRET") {
			continue
		}

		isLeaked := false
		lowerSrc := strings.ToLower(src)
		if strings.Contains(lowerSrc, "log") || strings.Contains(lowerSrc, "print") {
			leakRegex := regexp.MustCompile(`(?i)(?:log|print|console)\w*\s*\([^)]*` + regexp.QuoteMeta(varName) + `[^)]*\)`)
			if leakRegex.MatchString(src) {
				isLeaked = true
			}
		}

		if isLeaked {
			if e.isLineIgnored(f, lineNumber, "ENTROPY-SECRET-LEAK") {
				continue
			}
			return []core.CheckResult{{
				ID:         "ENTROPY-SECRET-LEAK",
				Name:       "Hardcoded Secret Leaked to Logger",
				Status:     core.Absent,
				File:       f.Path,
				Line:       lineNumber,
				Evidence:   fmt.Sprintf("Secret '%s' is not only hardcoded but also passed to a print/log function.", varName),
				Suggestion: "CRITICAL: Hardcoded secret is being logged. Remove the secret, remove the log statement, and rotate the credentials immediately.",
				Severity:   "CRITICAL",
			}}
		}

		return []core.CheckResult{{
			ID:         "ENTROPY-SECRET",
			Name:       "Hardcoded Secret Detected",
			Status:     core.Absent,
			File:       f.Path,
			Line:       lineNumber,
			Evidence:   fmt.Sprintf("Potential hardcoded secret in variable '%s': %s...", varName, val[:min(8, len(val))]),
			Suggestion: "CRITICAL: Hardcoded secret found in source code. Move to environment variables and rotate the value immediately.",
			Severity:   "CRITICAL",
		}}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (e *Engine) evaluateEntropyPattern(f *core.FileInfo) []core.CheckResult {
	var results []core.CheckResult

	content, err := f.GetContent()
	if err != nil {
		return results
	}
	s := string(content)

	isCodeFile := false
	exts := []string{".go", ".ts", ".tsx", ".js", ".jsx", ".py"}
	for _, ext := range exts {
		if strings.HasSuffix(f.Path, ext) {
			isCodeFile = true
			break
		}
	}
	isNodeModules := strings.Contains(f.Path, "node_modules")

	if isCodeFile && !isNodeModules {
		if matches := placeholderRegex.FindAllStringIndex(s, -1); len(matches) > 5 {
			results = append(results, core.CheckResult{
				ID: "ENTR-SLOP-01", Name: "Excessive AI Scaffolding Residue", Status: core.Absent,
				File: f.Path, Evidence: fmt.Sprintf("Found %d placeholder patterns.", len(matches)),
				Suggestion: "HIGH: Code contains high volume of scaffolding residue. Clean up placeholders.",
				Severity:   "HIGH",
			})
		}
	}

	if (f.Extension == "go" || f.Extension == "js" || f.Extension == "ts") && !strings.Contains(s, "err") && !strings.Contains(s, "catch") {
		if len(s) > 2000 {
			results = append(results, core.CheckResult{
				ID: "ENTR-FRAGILE", Name: "Brittle Logic Block", Status: core.Absent,
				File: f.Path, Evidence: "Large file with NO error handling keywords.",
				Suggestion: "HIGH: This file looks like an unstructured code dump. Implement robust error handling.",
				Severity:   "HIGH",
			})
		}
	}
	return results
}

func (e *Engine) evaluateFile(rule Rule, f *core.FileInfo) (string, int) {
	if rule.ExcludeTests && f.IsTest {
		return "", 0
	}
	if matchesExcludedPath(rule, f) {
		return "", 0
	}

	// 1. Handle logical Patterns (AND)
	if len(rule.Patterns) > 0 {
		var firstEvidence string
		var firstLine int
		for _, subRule := range rule.Patterns {
			ev, line := e.evaluateFile(subRule, f)
			if line == 0 {
				return "", 0 // One failed, all fail
			}
			if firstLine == 0 {
				firstEvidence, firstLine = ev, line
			}
		}
		return firstEvidence, firstLine
	}

	// 2. Handle logical PatternEither (OR)
	if len(rule.PatternEither) > 0 {
		for _, subRule := range rule.PatternEither {
			ev, line := e.evaluateFile(subRule, f)
			if line != 0 {
				return ev, line // One matched, we are good
			}
		}
		return "", 0
	}

	// 3. Handle logical PatternNot
	if rule.PatternNot != "" {
		notRule := rule
		notRule.Pattern = rule.PatternNot
		notRule.PatternNot = ""
		_, line := e.evaluateFile(notRule, f)
		if line != 0 {
			return "", 0 // Matched the negative pattern
		}
		// If we are just a PatternNot, we need something else to match positively?
		// Semgrep usually uses PatternNot alongside other patterns.
		// If it's alone, it's weird, but we'll return 1 to indicate "valid" if not matched.
		if rule.Pattern == "" {
			return "PatternNot validated: not present", 1
		}
	}

	// 4. Base Pattern Match
	var targetStr string
	var err error

	switch rule.Target {
	case "code", "code_only", "": // Default is code
		b, err2 := f.GetStrippedContent()
		err = err2
		targetStr = string(b)
	case "docs":
		b, err2 := f.GetDocsOnlyContent()
		err = err2
		targetStr = string(b)
	case "all":
		b, err2 := f.GetContent()
		err = err2
		targetStr = string(b)
	case "condition":
		if rule.Condition == "large_file" {
			b, _ := f.GetContent()
			lines := strings.Count(string(b), "\n")
			if lines > 1500 {
				return fmt.Sprintf("File %s has %d lines (God File)", f.Path, lines), 1
			}
		}
		return "", 0
	case "ast":
		tree, err2 := f.GetTree()
		b, err3 := f.GetContent()
		if err2 == nil && err3 == nil {
			lang, _ := ast.GetLanguage(f.Extension)
			matches, astErr := ast.MatchInTree(lang, tree, b, rule.Pattern)
			if astErr == nil && len(matches) > 0 {
				// Use the real line number from the matched AST node
				node := matches[0].Node
				if node != nil {
					return fmt.Sprintf("Rule %s triggered on %s", rule.ID, f.Path),
						int(node.StartPosition().Row) + 1 // tree-sitter rows are 0-indexed
				}
				return fmt.Sprintf("Rule %s triggered on %s", rule.ID, f.Path), 1
			}
		}
		if err2 != nil {
			err = err2
		} else {
			err = err3
		}
	default:
		// Fallback to code if unknown
		b, _ := f.GetStrippedContent()
		targetStr = string(b)
	}

	if err != nil {
		return "", 0
	}

	if token, ok := strings.CutPrefix(rule.Condition, "not_contains:"); ok &&
		strings.Contains(strings.ToLower(targetStr), strings.ToLower(token)) {
		return "", 0
	}

	// Pattern Match
	if rule.Target == "ast" {
		return "", 0
	}

	if rule.Pattern != "" {
		// Prefer regex if already compiled, or try literal match if very simple
		if rule.CompiledPattern != nil {
			loc := rule.CompiledPattern.FindStringIndex(targetStr)
			if loc != nil {
				lineNumber := strings.Count(targetStr[:loc[0]], "\n") + 1
				if e.isLineIgnored(f, lineNumber, rule.ID) {
					return "", 0
				}
				return fmt.Sprintf("Rule %s triggered on %s", rule.ID, f.Path), lineNumber
			}
		} else {
			// Literal match fallback
			idx := strings.Index(targetStr, rule.Pattern)
			if idx != -1 {
				lineNumber := strings.Count(targetStr[:idx], "\n") + 1
				if e.isLineIgnored(f, lineNumber, rule.ID) {
					return "", 0
				}
				return fmt.Sprintf("Rule %s triggered on %s", rule.ID, f.Path), lineNumber
			}
		}
	}

	return "", 0
}

func (e *Engine) isLineIgnored(f *core.FileInfo, line int, ruleID string) bool {
	content, err := f.GetContent()
	if err != nil {
		return false
	}

	lines := strings.Split(string(content), "\n")

	// Check for file-level suppression (anywhere in the first 10 lines)
	scanLimit := 10
	if len(lines) < scanLimit {
		scanLimit = len(lines)
	}
	for i := 0; i < scanLimit; i++ {
		if containsSuppression(lines[i], "aitriage-ignore-file") || containsSuppression(lines[i], "aitriage:ignore-file") {
			return true
		}
	}

	if line < 1 || line > len(lines) {
		return false
	}

	// Check same line and previous line for inline suppression
	checkLines := []int{line - 1} // Current line (0-indexed)
	if line > 1 {
		checkLines = append(checkLines, line-2) // Previous line
	}

	for _, lIdx := range checkLines {
		if lIdx < 0 || lIdx >= len(lines) {
			continue
		}
		l := lines[lIdx]
		// Support both syntaxes:
		//   // aitriage-ignore: RULE-ID (reason)
		//   // aitriage:ignore RULE-ID
		//   # aitriage-ignore-next-line: RULE-ID
		if containsSuppression(l, "aitriage-ignore") || containsSuppression(l, "aitriage:ignore") {
			if strings.Contains(l, "all") || strings.Contains(l, ruleID) {
				return true
			}
		}
	}

	return false
}

// containsSuppression checks if a line contains a suppression directive.
func containsSuppression(line, directive string) bool {
	return strings.Contains(line, directive)
}

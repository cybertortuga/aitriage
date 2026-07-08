package nfr

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed rules/*.yaml
var rulesFS embed.FS

var (
	cachedRules   []Rule
	loadRulesOnce sync.Once
	loadRulesErr  error
)

type Rule struct {
	ID            string   `yaml:"id"`
	Name          string   `yaml:"name"`
	Severity      string   `yaml:"severity"`
	Message       string   `yaml:"message"`
	Advice        string   `yaml:"advice"`
	Check         string   `yaml:"check"`   // "file_contains" | "file_exists" | "file_not_exists"
	Pattern       string   `yaml:"pattern"` // regex для file_contains
	Files         []string `yaml:"files"`   // glob паттерны файлов для проверки
	compiledRegex *regexp.Regexp
}

func getRules() ([]Rule, error) {
	loadRulesOnce.Do(func() {
		entries, err := rulesFS.ReadDir("rules")
		if err != nil {
			loadRulesErr = fmt.Errorf("cannot read embedded NFR rules: %w", err)
			return
		}

		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".yaml") && !strings.HasSuffix(e.Name(), ".yml") {
				continue
			}
			data, err := rulesFS.ReadFile("rules/" + e.Name())
			if err != nil {
				continue
			}
			var rules []Rule
			if err := yaml.Unmarshal(data, &rules); err != nil {
				continue
			}

			// Precompile regular expressions
			for i := range rules {
				if rules[i].Check == "file_contains" {
					re, err := regexp.Compile(rules[i].Pattern)
					if err == nil {
						rules[i].compiledRegex = re
					}
				}
			}

			cachedRules = append(cachedRules, rules...)
		}
	})
	return cachedRules, loadRulesErr
}

type NFRFinding struct {
	RuleID   string `json:"rule_id"`
	Name     string `json:"name"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Advice   string `json:"advice"`
}

// CheckNFR проверяет проект на соответствие NFR встроенным правилам
func CheckNFR(projectPath string) ([]NFRFinding, error) {
	allRules, err := getRules()
	if err != nil {
		return nil, err
	}

	var findings []NFRFinding
	for _, rule := range allRules {
		triggered, err := evaluateRule(projectPath, rule)
		if err != nil {
			continue
		}
		if triggered {
			findings = append(findings, NFRFinding{
				RuleID:   rule.ID,
				Name:     rule.Name,
				Severity: rule.Severity,
				Message:  rule.Message,
				Advice:   rule.Advice,
			})
		}
	}

	return findings, nil
}

func evaluateRule(projectPath string, rule Rule) (bool, error) {
	switch rule.Check {
	case "file_contains":
		var re *regexp.Regexp
		if rule.compiledRegex != nil {
			re = rule.compiledRegex
		} else {
			var err error
			re, err = regexp.Compile(rule.Pattern)
			if err != nil {
				return false, err
			}
		}

		found := false
		for _, glob := range rule.Files {
			matches, _ := filepath.Glob(filepath.Join(projectPath, "**", glob))
			matches2, _ := filepath.Glob(filepath.Join(projectPath, glob))
			matches = append(matches, matches2...)
			for _, path := range matches {
				data, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				if re.Match(data) {
					found = true
					break
				}
			}
		}
		return !found, nil // NFR нарушено если паттерн НЕ найден
	case "file_exists":
		_, err := os.Stat(filepath.Join(projectPath, rule.Pattern))
		return os.IsNotExist(err), nil // нарушено если файл НЕ существует
	case "file_not_exists":
		_, err := os.Stat(filepath.Join(projectPath, rule.Pattern))
		return err == nil, nil // нарушено если файл СУЩЕСТВУЕТ
	default:
		return false, nil
	}
}

// GetAllRulesAsText returns all NFR rules serialized as text for LLM consumption
func GetAllRulesAsText() string {
	allRules, err := getRules()
	if err != nil {
		return ""
	}

	var builder strings.Builder
	for _, r := range allRules {
		builder.WriteString(fmt.Sprintf("- [%s] %s (%s)\n  Advice: %s\n", r.ID, r.Name, r.Severity, r.Advice))
	}
	return builder.String()
}

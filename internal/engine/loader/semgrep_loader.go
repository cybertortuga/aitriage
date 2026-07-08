package loader

import (
	"fmt"
	"os"

	"github.com/cybertortuga/aitriage/internal/models"
	"gopkg.in/yaml.v3"
)

// LoadSemgrepRules loads rules from a YAML file in Semgrep format
func LoadSemgrepRules(path string) ([]models.Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	var raw struct {
		Rules []models.Rule `yaml:"rules"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	// Post-processing for each rule
	for i := range raw.Rules {
		raw.Rules[i] = processRule(raw.Rules[i])
	}

	return raw.Rules, nil
}

func processRule(r models.Rule) models.Rule {
	// 1. If Pattern is set, but it looks like a tree-sitter query, mark it as 'ast' target
	// Semgrep rules don't usually have 'target: ast' explicitly, they infer it.
	// For our MVP, we'll assume anything that isn't a simple regex or is in 'patterns' for certain languages is AST.
	// But to stay safe, let's look for Tree-sitter specific syntax or the presence of $VAR.

	// Default to 'ast' if Pattern looks like a TS query (contains parens and captures)
	if r.Pattern != "" {
		if (len(r.Pattern) > 2 && r.Pattern[0] == '(') || (fmt.Sprintf("%v", r.Pattern)) == "..." {
			r.Target = "ast"
		}
	}

	// Recursively process nested rules
	for i := range r.Patterns {
		r.Patterns[i] = processRule(r.Patterns[i])
		// Propagate target if parent is AST
		if r.Target == "ast" {
			r.Patterns[i].Target = "ast"
		}
	}

	for i := range r.PatternEither {
		r.PatternEither[i] = processRule(r.PatternEither[i])
		if r.Target == "ast" {
			r.PatternEither[i].Target = "ast"
		}
	}

	return r
}

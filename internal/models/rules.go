package models

import "regexp"

type Rule struct {
	ID           string   `yaml:"id" json:"id"`
	Name         string   `yaml:"name" json:"name"`
	Stack        string   `yaml:"stack" json:"stack"`
	Extensions   []string `yaml:"extensions" json:"extensions"`
	Target       string   `yaml:"target" json:"target"`
	Pattern      string   `yaml:"pattern" json:"pattern"`
	Condition    string   `yaml:"condition" json:"condition"`
	Files        []string `yaml:"files" json:"files"`
	ExcludeTests bool     `yaml:"exclude_tests" json:"exclude_tests"`
	ExcludePaths []string `yaml:"exclude_paths" json:"exclude_paths"` // path fragments to skip (e.g. "seed", "migration")
	Suggestion   string   `yaml:"suggestion" json:"suggestion"`
	Severity     string   `yaml:"severity" json:"severity"`
	Message      string   `yaml:"message" json:"message"`
	Languages    []string `yaml:"languages" json:"languages"`

	// Semgrep-like logical patterns
	Patterns         []Rule `yaml:"patterns" json:"patterns,omitempty"`
	PatternEither    []Rule `yaml:"pattern-either" json:"pattern_either,omitempty"`
	PatternNot       string `yaml:"pattern-not" json:"pattern_not,omitempty"`
	PatternInside    string `yaml:"pattern-inside" json:"pattern_inside,omitempty"`
	PatternNotInside string `yaml:"pattern-not-inside" json:"pattern_not_inside,omitempty"`

	CompiledPattern *regexp.Regexp `yaml:"-" json:"-"`
}

type Ruleset struct {
	Rules []Rule `yaml:"rules" json:"rules"`
}

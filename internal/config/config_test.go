package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_CustomRules(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aitriage-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configContent := `
ignore:
  rules: ["ENTR-01"]
  paths: ["vendor/"]
rules:
  - id: "CUSTOM-01"
    name: "Custom Rule"
    stack: "universal"
    target: "code"
    pattern: "TODO: custom"
    suggestion: "Fix it"
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".aitriage.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := LoadConfig(tmpDir)
	if cfg == nil {
		t.Fatal("Config should not be nil")
	}

	if len(cfg.Ignore.Rules) != 1 || cfg.Ignore.Rules[0] != "ENTR-01" {
		t.Errorf("Ignore rules not loaded correctly: %v", cfg.Ignore.Rules)
	}

	if len(cfg.CustomRules) != 1 || cfg.CustomRules[0].ID != "CUSTOM-01" {
		t.Errorf("Custom rules not loaded correctly: %v", cfg.CustomRules)
	}

	if cfg.CustomRules[0].Pattern != "TODO: custom" {
		t.Errorf("Custom rule pattern mismatch: %s", cfg.CustomRules[0].Pattern)
	}
}

func TestIsRuleIgnored(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		ruleID   string
		expected bool
	}{
		{
			name:     "Empty Config",
			config:   &Config{},
			ruleID:   "RULE-01",
			expected: false,
		},
		{
			name: "Rule is ignored",
			config: &Config{
				Ignore: IgnoreConfig{
					Rules: []string{"RULE-01", "RULE-02"},
				},
			},
			ruleID:   "RULE-02",
			expected: true,
		},
		{
			name: "Rule is not ignored",
			config: &Config{
				Ignore: IgnoreConfig{
					Rules: []string{"RULE-01", "RULE-03"},
				},
			},
			ruleID:   "RULE-02",
			expected: false,
		},
		{
			name: "Empty rule ID is ignored if listed",
			config: &Config{
				Ignore: IgnoreConfig{
					Rules: []string{"RULE-01", ""},
				},
			},
			ruleID:   "",
			expected: true,
		},
		{
			name: "Empty rule ID is not ignored if not listed",
			config: &Config{
				Ignore: IgnoreConfig{
					Rules: []string{"RULE-01"},
				},
			},
			ruleID:   "",
			expected: false,
		},
		{
			name: "Multiple ignored rules in config, matches last",
			config: &Config{
				Ignore: IgnoreConfig{
					Rules: []string{"RULE-01", "RULE-02", "RULE-03"},
				},
			},
			ruleID:   "RULE-03",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsRuleIgnored(tt.ruleID)
			if result != tt.expected {
				t.Errorf("Expected IsRuleIgnored(%q) to be %v, got %v", tt.ruleID, tt.expected, result)
			}
		})
	}
}

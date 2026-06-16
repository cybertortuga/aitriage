package config

import (
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Ignore      IgnoreConfig `yaml:"ignore,omitempty"`
	CustomRules []CustomRule `yaml:"rules,omitempty"`
	LLM         LLMConfig    `yaml:"llm,omitempty"`
	StrictMode  bool         `yaml:"strict_mode,omitempty"` // Fail on ANY finding
	FailScore   int          `yaml:"fail_score,omitempty"`  // Fail if SecurityScore < this value
}

type LLMConfig struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	APIKey   string `yaml:"api_key"`
	BaseURL  string `yaml:"base_url"`
	Timeout  int    `yaml:"timeout"`
}

type CustomRule struct {
	ID         string   `yaml:"id"`
	Name       string   `yaml:"name"`
	Stack      string   `yaml:"stack"`
	Extensions []string `yaml:"extensions,omitempty"`
	Target     string   `yaml:"target"`
	Pattern    string   `yaml:"pattern"`
	Condition  string   `yaml:"condition,omitempty"`
	Files      []string `yaml:"files,omitempty"`
	Suggestion string   `yaml:"suggestion"`
	Severity   string   `yaml:"severity,omitempty"`
}

type IgnoreConfig struct {
	Rules []string `yaml:"rules,omitempty"`
	Paths []string `yaml:"paths,omitempty"`
}

// LoadConfig looks for aitriage.yaml or .aitriage.yaml in the given root string.
// After loading, it applies env var overrides for LLM config (env > yaml).
func LoadConfig(rootPath string) *Config {
	possibleFiles := []string{"aitriage.yaml", ".aitriage.yaml", "aitriage.yml", ".aitriage.yml"}

	var configData []byte
	var err error
	var found bool

	for _, file := range possibleFiles {
		fullPath := filepath.Join(rootPath, file)
		configData, err = os.ReadFile(fullPath)
		if err == nil {
			found = true
			break
		}
	}

	cfg := &Config{}
	if found {
		// Expand $VAR and ${VAR} patterns before parsing.
		// This allows users to reference env vars in config: api_key: $GEMINI_API_KEY
		expandedData := os.ExpandEnv(string(configData))
		err = yaml.Unmarshal([]byte(expandedData), cfg)
		if err != nil {
			slog.Warn("Failed to parse configuration file", "error", err)
		}
	}

	// Auto-load custom generated rules from .aitriage/custom_rules.yaml
	customRulesPath := filepath.Join(rootPath, ".aitriage", "custom_rules.yaml")
	if customData, err := os.ReadFile(customRulesPath); err == nil {
		var customRules []CustomRule
		if err := yaml.Unmarshal(customData, &customRules); err == nil {
			cfg.CustomRules = append(cfg.CustomRules, customRules...)
		} else {
			slog.Warn("Failed to parse custom_rules.yaml", "error", err)
		}
	}

	// Auto-detect LLM provider from env vars if not set in YAML.
	// Priority: GEMINI_API_KEY > ANTHROPIC_API_KEY > OPENAI_API_KEY
	applyEnvLLMConfig(&cfg.LLM)

	return cfg
}

// applyEnvLLMConfig fills in missing LLM config fields from environment variables.
// CLI flags applied later in agent.go will still override these.
func applyEnvLLMConfig(llm *LLMConfig) {
	// API key env overrides (only if not set in yaml)
	if llm.APIKey == "" {
		for _, env := range []string{"GEMINI_API_KEY", "GOOGLE_API_KEY", "ANTHROPIC_API_KEY", "OPENAI_API_KEY"} {
			if v := os.Getenv(env); v != "" {
				llm.APIKey = v
				// Auto-set provider if not specified
				if llm.Provider == "" {
					switch env {
					case "GEMINI_API_KEY", "GOOGLE_API_KEY":
						llm.Provider = "gemini"
					case "ANTHROPIC_API_KEY":
						llm.Provider = "anthropic"
					case "OPENAI_API_KEY":
						llm.Provider = "openai"
					}
				}
				break
			}
		}
	}
	// Explicit env overrides for provider/model/base_url
	if v := os.Getenv("AITRIAGE_LLM_PROVIDER"); v != "" {
		llm.Provider = v
	}
	if v := os.Getenv("AITRIAGE_LLM_MODEL"); v != "" {
		llm.Model = v
	}
	if v := os.Getenv("AITRIAGE_LLM_BASE_URL"); v != "" {
		llm.BaseURL = v
	}
}

func (c *Config) IsRuleIgnored(ruleID string) bool {
	for _, id := range c.Ignore.Rules {
		if id == ruleID {
			return true
		}
	}
	return false
}

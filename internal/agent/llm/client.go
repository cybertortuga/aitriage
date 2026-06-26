package llm

import "context"

// Message represents a single chat message.
type Message struct {
	Role    string `json:"role"` // "system" | "user" | "assistant"
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Client is the interface for any LLM provider.
type Client interface {
	// Chat sends a list of messages and returns the model's response and usage stats.
	Chat(ctx context.Context, messages []Message) (string, Usage, error)
}

// Config holds the LLM provider configuration from .aitriage.yaml or env.
type Config struct {
	Provider        string `yaml:"provider"` // "gemini" | "anthropic" | "openai" | "ollama" | "groq"
	Model           string `yaml:"model"`
	APIKey          string `yaml:"api_key"`
	BaseURL         string `yaml:"base_url"`         // для ollama и openai-compatible
	Timeout         int    `yaml:"timeout"`           // секунды, default 120
	DisableThinking bool   `yaml:"disable_thinking"` // Send thinking:{type:disabled} for reasoning models
}

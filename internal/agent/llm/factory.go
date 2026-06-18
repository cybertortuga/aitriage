package llm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/openai/openai-go"
	openai_option "github.com/openai/openai-go/option"
)

const geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/openai/"

const noLLMConfiguredMsg = `no LLM provider configured

Set one of these environment variables:
  export GEMINI_API_KEY=your-key     # Google Gemini (recommended)
  export ANTHROPIC_API_KEY=your-key  # Anthropic Claude
  export OPENAI_API_KEY=your-key     # OpenAI GPT

Or add to .aitriage.yaml:
  llm:
    provider: gemini
    api_key: your-key`

// openAIClient wraps the official openai-go SDK.
type openAIClient struct {
	client openai.Client
	cfg    Config
}

func (c *openAIClient) Chat(ctx context.Context, messages []Message) (string, Usage, error) {
	var oaiMessages []openai.ChatCompletionMessageParamUnion
	for _, m := range messages {
		switch m.Role {
		case "system":
			oaiMessages = append(oaiMessages, openai.SystemMessage(m.Content))
		case "user":
			oaiMessages = append(oaiMessages, openai.UserMessage(m.Content))
		case "assistant":
			oaiMessages = append(oaiMessages, openai.AssistantMessage(m.Content))
		}
	}

	model := c.cfg.Model
	if model == "" {
		model = string(openai.ChatModelGPT4o)
	}

	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: oaiMessages,
		Model:    openai.ChatModel(model),
	})
	if err != nil {
		return "", Usage{}, fmt.Errorf("openai chat error: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", Usage{}, fmt.Errorf("empty response from openai")
	}

	usage := Usage{}
	if completion.Usage.TotalTokens > 0 {
		usage.PromptTokens = int(completion.Usage.PromptTokens)
		usage.CompletionTokens = int(completion.Usage.CompletionTokens)
		usage.TotalTokens = int(completion.Usage.TotalTokens)
	}

	return completion.Choices[0].Message.Content, usage, nil
}

// anthropicClientWrapper wraps the official anthropic-sdk-go SDK.
type anthropicClientWrapper struct {
	client anthropic.Client
	cfg    Config
}

func (c *anthropicClientWrapper) Chat(ctx context.Context, messages []Message) (string, Usage, error) {
	var systemBlocks []anthropic.TextBlockParam
	var anthropicMessages []anthropic.MessageParam

	for _, m := range messages {
		switch m.Role {
		case "system":
			systemBlocks = append(systemBlocks, anthropic.TextBlockParam{
				Text: m.Content,
				Type: "text",
			})
		case "user":
			anthropicMessages = append(anthropicMessages,
				anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		case "assistant":
			anthropicMessages = append(anthropicMessages,
				anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		}
	}

	model := c.cfg.Model
	if model == "" {
		model = anthropic.ModelClaudeSonnet4_5
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 8192,
		Messages:  anthropicMessages,
	}

	if len(systemBlocks) > 0 {
		params.System = systemBlocks
	}

	message, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return "", Usage{}, fmt.Errorf("anthropic chat error: %w", err)
	}

	if len(message.Content) == 0 {
		return "", Usage{}, fmt.Errorf("empty response from anthropic")
	}

	usage := Usage{
		PromptTokens:     int(message.Usage.InputTokens),
		CompletionTokens: int(message.Usage.OutputTokens),
		TotalTokens:      int(message.Usage.InputTokens + message.Usage.OutputTokens),
	}

	return message.Content[0].Text, usage, nil
}

// NewClient создаёт LLM клиент нужного провайдера на основе конфига.
// All clients are automatically wrapped with RetryClient (3 retries,
// exponential backoff) to handle transient 429/5xx/network errors.
func NewClient(cfg Config) (Client, error) {
	var inner Client
	var err error

	switch cfg.Provider {
	case "anthropic":
		if cfg.APIKey == "" {
			return nil, errors.New(noLLMConfiguredMsg)
		}
		opts := []option.RequestOption{option.WithAPIKey(cfg.APIKey)}
		if cfg.Timeout > 0 {
			opts = append(opts, option.WithRequestTimeout(time.Duration(cfg.Timeout)*time.Second))
		}
		inner = &anthropicClientWrapper{
			client: anthropic.NewClient(opts...),
			cfg:    cfg,
		}

	case "gemini":
		// Gemini uses the OpenAI-compatible API.
		// API key is resolved by config.LoadConfig() from env vars (GEMINI_API_KEY / GOOGLE_API_KEY).
		if cfg.APIKey == "" {
			return nil, errors.New(noLLMConfiguredMsg)
		}
		baseURL := geminiBaseURL
		if cfg.BaseURL != "" {
			baseURL = cfg.BaseURL
		}
		model := cfg.Model
		if model == "" {
			model = "gemini-2.5-flash"
		}
		opts := []openai_option.RequestOption{
			openai_option.WithAPIKey(cfg.APIKey),
			openai_option.WithBaseURL(baseURL),
		}
		if cfg.Timeout > 0 {
			opts = append(opts, openai_option.WithRequestTimeout(time.Duration(cfg.Timeout)*time.Second))
		}
		inner = &openAIClient{
			client: openai.NewClient(opts...),
			cfg:    Config{Provider: "gemini", Model: model, APIKey: cfg.APIKey, BaseURL: baseURL, Timeout: cfg.Timeout},
		}

	case "openai", "ollama", "groq":
		if cfg.APIKey == "" && cfg.Provider == "openai" {
			return nil, errors.New(noLLMConfiguredMsg)
		}
		opts := []openai_option.RequestOption{}
		if cfg.APIKey != "" {
			opts = append(opts, openai_option.WithAPIKey(cfg.APIKey))
		}
		if cfg.BaseURL != "" {
			opts = append(opts, openai_option.WithBaseURL(cfg.BaseURL))
		}
		if cfg.Timeout > 0 {
			opts = append(opts, openai_option.WithRequestTimeout(time.Duration(cfg.Timeout)*time.Second))
		}
		inner = &openAIClient{
			client: openai.NewClient(opts...),
			cfg:    cfg,
		}

	case "":
		return nil, errors.New(noLLMConfiguredMsg)

	default:
		return nil, fmt.Errorf("unknown LLM provider: %q\nSupported: gemini, anthropic, openai, ollama, groq", cfg.Provider)
	}

	if err != nil {
		return nil, err
	}

	// Wrap with retry logic: 3 retries with exponential backoff.
	return NewRetryClient(inner, 3), nil
}

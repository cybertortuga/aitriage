package llm

import (
	"encoding/json"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go"
)

func TestUsageFromOpenAIExtractsCachedPromptTokens(t *testing.T) {
	var raw openai.CompletionUsage
	if err := json.Unmarshal([]byte(`{
		"prompt_tokens": 100,
		"completion_tokens": 20,
		"total_tokens": 120,
		"prompt_tokens_details": {"cached_tokens": 64}
	}`), &raw); err != nil {
		t.Fatalf("unmarshal openai usage: %v", err)
	}

	got := usageFromOpenAI(raw)
	if got.PromptTokens != 100 || got.CompletionTokens != 20 || got.TotalTokens != 120 {
		t.Fatalf("usage totals = %+v, want 100/20/120", got)
	}
	if got.CachedPromptTokens != 64 || !got.CacheTelemetryReported {
		t.Fatalf("cache telemetry = %+v, want cached prompt tokens reported", got)
	}
}

func TestUsageFromOpenAINotesMissingCacheTelemetry(t *testing.T) {
	var raw openai.CompletionUsage
	if err := json.Unmarshal([]byte(`{
		"prompt_tokens": 100,
		"completion_tokens": 20,
		"total_tokens": 120
	}`), &raw); err != nil {
		t.Fatalf("unmarshal openai usage: %v", err)
	}

	got := usageFromOpenAI(raw)
	if got.CacheTelemetryReported || got.CachedPromptTokens != 0 {
		t.Fatalf("cache telemetry = %+v, want provider_did_not_report", got)
	}
}

func TestUsageFromAnthropicExtractsCacheReadAndCreationTokens(t *testing.T) {
	var raw anthropic.Usage
	if err := json.Unmarshal([]byte(`{
		"input_tokens": 100,
		"output_tokens": 20,
		"cache_creation_input_tokens": 30,
		"cache_read_input_tokens": 40
	}`), &raw); err != nil {
		t.Fatalf("unmarshal anthropic usage: %v", err)
	}

	got := usageFromAnthropic(raw)
	if got.PromptTokens != 100 || got.CompletionTokens != 20 || got.TotalTokens != 120 {
		t.Fatalf("usage totals = %+v, want 100/20/120", got)
	}
	if got.CacheCreationInputTokens != 30 || got.CacheReadInputTokens != 40 || !got.CacheTelemetryReported {
		t.Fatalf("cache telemetry = %+v, want cache creation/read tokens reported", got)
	}
}

package graph

import (
	"context"
	"testing"

	"github.com/cybertortuga/aitriage/internal/agent/llm"
)

type stageUsageMockLLM struct {
	usage llm.Usage
}

func (m stageUsageMockLLM) Chat(ctx context.Context, messages []llm.Message) (string, llm.Usage, error) {
	return "{}", m.usage, nil
}

func TestStageUsageClientDetectsTriageStages(t *testing.T) {
	state := &AgentState{}
	client := trackTriageLLMStages(state, stageUsageMockLLM{
		usage: llm.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15, CachedPromptTokens: 3, CacheTelemetryReported: true},
	})

	_, _, _ = client.Chat(context.Background(), []llm.Message{{Role: "system", Content: "Threat Model & Finding Classification"}})
	_, _, _ = client.Chat(context.Background(), []llm.Message{{Role: "system", Content: "Finding Classification"}})

	if got := state.StageUsage[usageStageThreatModel]; got.TotalTokens != 15 || got.CachedPromptTokens != 3 || !got.CacheTelemetryReported {
		t.Fatalf("threat-model stage usage = %+v", got)
	}
	if got := state.StageUsage[usageStageClassification]; got.TotalTokens != 15 || got.CachedPromptTokens != 3 || !got.CacheTelemetryReported {
		t.Fatalf("classification stage usage = %+v", got)
	}
}

func TestFormatLLMUsageIncludesCacheTelemetry(t *testing.T) {
	got := formatLLMUsage(llm.Usage{
		PromptTokens:             100,
		CompletionTokens:         25,
		TotalTokens:              140,
		CachedPromptTokens:       60,
		CacheCreationInputTokens: 10,
		CacheReadInputTokens:     20,
		CacheTelemetryReported:   true,
	})
	want := "140 total · 100 prompt · 25 completion · 15 reasoning/other · cache telemetry: 60 cached prompt, 10 cache creation, 20 cache read"
	if got != want {
		t.Fatalf("formatLLMUsage() = %q, want %q", got, want)
	}
}

func TestFormatLLMUsageMarksProviderDidNotReportCacheTelemetry(t *testing.T) {
	got := formatLLMUsage(llm.Usage{PromptTokens: 100, CompletionTokens: 25, TotalTokens: 125})
	want := "125 total · 100 prompt · 25 completion · cache telemetry: provider_did_not_report"
	if got != want {
		t.Fatalf("formatLLMUsage() = %q, want %q", got, want)
	}
}

package mcp

import (
	"strings"
	"testing"
)

func TestContainerBundleRequirementAppliesToStartAndVerification(t *testing.T) {
	t.Setenv("AITRIAGE_RUNTIME", "container")
	if !containerBundleRequired() {
		t.Fatal("container MCP runs, including verification, must fail closed on incomplete scanner coverage")
	}
	t.Setenv("AITRIAGE_RUNTIME", "")
	if containerBundleRequired() {
		t.Fatal("native development mode must not masquerade as the verified container bundle")
	}
}

func TestServerInstructionsEnforceApprovalBeforeEdits(t *testing.T) {
	instructions := serverInstructions(true)
	for _, required := range []string{
		"FIRST action MUST be `aitriage_run_approve`",
		"BEFORE planning a patch, editing any file",
		"status=fixing and fix_context",
		"never approve retroactively",
		"project change made AFTER the recorded approval",
	} {
		if !strings.Contains(instructions, required) {
			t.Errorf("server instructions missing %q", required)
		}
	}
}

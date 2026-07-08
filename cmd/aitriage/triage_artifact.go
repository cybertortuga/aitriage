package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cybertortuga/aitriage/internal/agent/graph"
)

func writeTriageArtifact(path string, state *graph.AgentState) error {
	artifact, err := graph.BuildTriageFindingsArtifact(state)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal triage artifact: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write triage artifact to %s: %w", path, err)
	}
	return nil
}

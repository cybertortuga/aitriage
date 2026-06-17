package agentcontext

import (
	"fmt"
	"strings"

	"github.com/cybertortuga/aitriage/internal/agent/architect"
)

// RepoContext holds the complete gathered repository context for LLM consumption.
type RepoContext struct {
	ProjectTree  string    `json:"project_tree"`
	KeyFiles     []KeyFile `json:"key_files"`
	Stack        string    `json:"stack"`
	Architecture string    `json:"architecture"`
}

// BuildRepoContext gathers the full repository context from disk.
// This is called once before all LLM steps and stored in AgentState.
func BuildRepoContext(projectPath string) *RepoContext {
	ctx := &RepoContext{}

	// 1. Project tree (depth 4, max 300 files).
	ctx.ProjectTree = BuildProjectTree(projectPath, 4, 300)

	// 2. Key files (manifests, configs, entrypoints, security files).
	ctx.KeyFiles = ReadKeyFiles(projectPath, 8192)

	// 3. Stack detection (reuse architect component detection for now).
	components := architect.DetectComponents(projectPath)
	if len(components) > 0 {
		var stackParts []string
		for _, c := range components {
			stackParts = append(stackParts, fmt.Sprintf("%s (%s)", c.Name, c.Type))
		}
		ctx.Stack = strings.Join(stackParts, ", ")
	}

	// 4. Architecture diagram (Mermaid).
	ctx.Architecture, _ = architect.GenerateMermaidDiagram(projectPath)

	return ctx
}

// FormatForLLM renders the repo context as a text block suitable for LLM prompts.
// tokenBudget is an approximate character limit (1 token ≈ 4 chars).
// If the full context exceeds the budget, lower-priority sections are trimmed.
func (rc *RepoContext) FormatForLLM(tokenBudget int) string {
	if rc == nil {
		return ""
	}

	charBudget := tokenBudget * 4
	if charBudget <= 0 {
		charBudget = 40000 // Default ~10K tokens.
	}

	var sb strings.Builder

	// Section 1: Project structure (high priority).
	if rc.ProjectTree != "" {
		sb.WriteString("## Repository Structure\n```\n")
		tree := rc.ProjectTree
		if len(tree) > 3000 {
			tree = tree[:3000] + "\n... (truncated)"
		}
		sb.WriteString(tree)
		sb.WriteString("\n```\n\n")
	}

	// Section 2: Stack & Architecture (high priority).
	if rc.Stack != "" {
		sb.WriteString("## Detected Stack\n")
		sb.WriteString(rc.Stack + "\n\n")
	}
	if rc.Architecture != "" {
		sb.WriteString("## Architecture\n```mermaid\n")
		sb.WriteString(rc.Architecture)
		sb.WriteString("\n```\n\n")
	}

	// Section 3: Key files (medium priority — trim if over budget).
	if len(rc.KeyFiles) > 0 {
		sb.WriteString("## Key Project Files\n\n")

		// Group by role for cleaner output.
		roleOrder := []string{"manifest", "entrypoint", "security", "config"}
		roleLabels := map[string]string{
			"manifest":   "Package Manifests",
			"entrypoint": "Application Entrypoints",
			"security":   "Auth / Security / Routing",
			"config":     "Configuration",
		}

		for _, role := range roleOrder {
			var filesForRole []KeyFile
			for _, kf := range rc.KeyFiles {
				if kf.Role == role {
					filesForRole = append(filesForRole, kf)
				}
			}
			if len(filesForRole) == 0 {
				continue
			}

			sb.WriteString("### " + roleLabels[role] + "\n\n")
			for _, kf := range filesForRole {
				// Check budget before adding more files.
				if sb.Len() > charBudget {
					sb.WriteString("... (remaining files omitted due to context budget)\n")
					return sb.String()
				}

				sb.WriteString("#### `" + kf.Path + "`\n```\n")
				content := kf.Content
				// Per-file cap to keep things reasonable.
				if len(content) > 4000 {
					content = content[:4000] + "\n... (truncated)"
				}
				sb.WriteString(content)
				sb.WriteString("\n```\n\n")
			}
		}
	}

	return sb.String()
}

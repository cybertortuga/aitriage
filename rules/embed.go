// Package rules provides the embedded rules filesystem for the AITriage engine.
// All YAML rule files in this directory and its subdirectories are the single
// source of truth for the built-in rules. The engine loads them at startup.
package rules

import "embed"

// FS contains all rule YAML files embedded at compile time.
// Subdirectories map to technology stacks (e.g., django/, express/, golang/).
//
//go:embed all:*
var FS embed.FS

package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Profile controls which MCP tools are exposed to a client.
//
// The default (ProfileFull) preserves historical behaviour: every tool is
// registered, including the mutating securecoder_ignore. The safe profile
// is the safe-by-default mode for low-trust or automated installs: it exposes
// only read-only scan/analysis tools and refuses to register any mutating tool.
type Profile string

const (
	// ProfileFull exposes all tools, including mutating ones. Backward-compatible default.
	ProfileFull Profile = "full"
	// ProfileSafe exposes only read-only scan/analysis tools. Mutating tools
	// (securecoder_ignore) are not registered at all.
	ProfileSafe Profile = "safe"
)

// ParseProfile normalises a user-supplied profile string. An empty string maps
// to ProfileFull so callers that do not opt in keep the historical behaviour.
func ParseProfile(s string) (Profile, error) {
	switch Profile(strings.ToLower(strings.TrimSpace(s))) {
	case "", ProfileFull:
		return ProfileFull, nil
	case ProfileSafe:
		return ProfileSafe, nil
	default:
		return "", fmt.Errorf("unknown profile %q (supported: full, safe)", s)
	}
}

// allowsMutation reports whether mutating tools may be registered under this profile.
func (p Profile) allowsMutation() bool { return p != ProfileSafe }

// Config configures a Server.
type Config struct {
	// Profile selects the exposed tool set. Empty means ProfileFull.
	Profile Profile
	// ScanRoot, when non-empty, confines every path argument to this directory
	// (see PathGuard). Empty means no confinement — the historical behaviour.
	ScanRoot string
}

// PathGuard confines tool path arguments to a single resolved root directory.
//
// A nil *PathGuard performs no confinement and returns inputs unchanged, so the
// full profile (which builds no guard) behaves exactly as before. When a root is
// configured, every path is resolved through filepath.EvalSymlinks and rejected
// unless the fully resolved result stays inside the root — this blocks both
// "../" traversal and symlinks that point outside the project.
type PathGuard struct {
	root string // absolute, symlink-resolved, cleaned
}

// NewPathGuard resolves root to an absolute, symlink-free directory path.
func NewPathGuard(root string) (*PathGuard, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("scan root is empty")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve scan root: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, fmt.Errorf("scan root %q is not accessible: %w", root, err)
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("scan root %q is not a directory", root)
	}
	return &PathGuard{root: resolved}, nil
}

// Root returns the confined root directory, or "" for a nil guard.
func (g *PathGuard) Root() string {
	if g == nil {
		return ""
	}
	return g.root
}

// Resolve validates input against the guard's root and returns a safe absolute
// path to hand to a scanner.
//
//   - nil guard: passthrough — returns input unchanged.
//   - empty input or ".": resolves to the root itself.
//   - relative input: joined onto the root.
//   - absolute input: taken as-is, then confined.
//
// The returned path is fully symlink-resolved and guaranteed to live inside the
// root. Any escape (traversal or symlink) is an error and no filesystem access
// is performed by the caller.
func (g *PathGuard) Resolve(input string) (string, error) {
	if g == nil {
		return input, nil
	}
	input = strings.TrimSpace(input)

	var candidate string
	switch {
	case input == "" || input == ".":
		candidate = g.root
	case filepath.IsAbs(input):
		candidate = filepath.Clean(input)
	default:
		candidate = filepath.Join(g.root, input)
	}

	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", fmt.Errorf("path %q not found or inaccessible within scan root", input)
	}
	if !within(g.root, resolved) {
		return "", fmt.Errorf("path %q escapes the allowed scan root (%s)", input, g.root)
	}
	return resolved, nil
}

// within reports whether p is root itself or a descendant of root. Both
// arguments must already be cleaned, absolute, symlink-resolved paths.
func within(root, p string) bool {
	if p == root {
		return true
	}
	return strings.HasPrefix(p, root+string(os.PathSeparator))
}

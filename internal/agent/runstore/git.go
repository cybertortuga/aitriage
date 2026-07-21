package runstore

import (
	"os/exec"
	"strings"
)

// CollectGitInfo gathers safe VCS context for a run manifest: the current
// commit, whether the working tree is dirty, and a content-free tree
// fingerprint. It is fail-soft — outside a git repository every field is simply
// empty, which is a valid manifest state.
func CollectGitInfo(root string) GitInfo {
	var info GitInfo
	if out, err := gitOutput(root, "rev-parse", "HEAD"); err == nil {
		info.Commit = strings.TrimSpace(out)
	}
	if out, err := gitOutput(root, "status", "--porcelain"); err == nil {
		info.Dirty = strings.TrimSpace(out) != ""
	}
	if fp, err := TreeFingerprint(root); err == nil {
		info.TreeFingerprint = fp
	}
	return info
}

func gitOutput(root string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.Output()
	return string(out), err
}

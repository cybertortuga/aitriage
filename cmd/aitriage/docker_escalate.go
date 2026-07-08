package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	dockerImageLocal    = "aitriage:local"
	dockerImageRegistry = "ghcr.io/cybertortuga/aitriage:latest"
)

// needsDocker checks if external SAST tools are missing.
func needsDocker() bool {
	for _, t := range []string{"semgrep", "trivy", "gitleaks", "bandit"} {
		if _, err := exec.LookPath(t); err != nil {
			return true
		}
	}
	return false
}

// hasDocker checks if Docker is available and running.
func hasDocker() bool {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// isInsideDocker returns true if we're already in a container.
func isInsideDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		s := string(data)
		if strings.Contains(s, "docker") || strings.Contains(s, "containerd") {
			return true
		}
	}
	return false
}

// dockerEscalate transparently re-launches the command inside Docker.
// The user never needs to know Docker exists — it's an implementation detail.
// Returns true if escalation happened (caller should exit).
func dockerEscalate(projectPath string) bool {
	if isInsideDocker() || !needsDocker() || !hasDocker() {
		if !isInsideDocker() && needsDocker() && !hasDocker() {
			printMissingToolsWarning()
		}
		return false
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return false
	}

	fmt.Fprintf(os.Stderr, "\n\033[38;2;0;245;255m⚡\033[0m \033[38;2;132;148;149mExternal scanners not found locally\033[0m\n")
	fmt.Fprintf(os.Stderr, "\033[38;2;0;245;255m🐳\033[0m \033[1m\033[38;2;220;228;228mLaunching full security audit in Docker...\033[0m\n\n")

	// 1) Try local image (built via make docker-build)
	if tryRunInDocker(dockerImageLocal, absPath) {
		return true
	}

	// 2) Try registry image
	if tryRunInDocker(dockerImageRegistry, absPath) {
		return true
	}

	// 3) Build locally from Dockerfile
	if tryLocalBuild(absPath) {
		return true
	}

	fmt.Fprintf(os.Stderr, "\033[38;2;231;196;39m⚠\033[0m  \033[38;2;132;148;149mDocker escalation failed, falling back to local scan\033[0m\n\n")
	return false
}

// tryRunInDocker attempts to pull and run the published image.
func tryRunInDocker(image, absPath string) bool {
	args := buildDockerArgs(image, absPath)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run() == nil
}

// tryLocalBuild builds from local Dockerfile and runs.
func tryLocalBuild(absPath string) bool {
	// Find Dockerfile — check current dir and project dir
	dockerfileDir := ""
	for _, dir := range []string{".", absPath} {
		if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err == nil {
			dockerfileDir = dir
			break
		}
	}
	if dockerfileDir == "" {
		return false
	}

	fmt.Fprintf(os.Stderr, "\033[38;2;0;245;255m🔨\033[0m \033[38;2;220;228;228mBuilding Docker image locally...\033[0m\n")
	build := exec.Command("docker", "build", "-t", "aitriage:local", dockerfileDir)
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return false
	}

	args := buildDockerArgs("aitriage:local", absPath)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run() == nil
}

// buildDockerArgs constructs docker run arguments.
func buildDockerArgs(image, absPath string) []string {
	args := []string{
		"run", "--rm", "-it",
		"-e", "TERM=xterm-256color",
		"-v", absPath + ":/project:ro",
	}

	// Forward API keys securely without exposing them in process arguments
	for _, key := range []string{"GEMINI_API_KEY", "OPENAI_API_KEY"} {
		if os.Getenv(key) != "" {
			args = append(args, "-e", key)
		}
	}

	// Image + rebuild original CLI args with /project substitution
	args = append(args, image)
	args = append(args, rebuildArgs()...)
	return args
}

// rebuildArgs reconstructs CLI args, replacing project path with /project.
func rebuildArgs() []string {
	osArgs := os.Args[1:]
	result := make([]string, 0, len(osArgs))
	for _, arg := range osArgs {
		if arg == "." || arg == "./" {
			result = append(result, "/project")
		} else if filepath.IsAbs(arg) {
			// Absolute paths get replaced too
			result = append(result, "/project")
		} else if !strings.HasPrefix(arg, "-") && arg != "scan" && arg != "agent" && arg != "web" {
			result = append(result, "/project")
		} else {
			result = append(result, arg)
		}
	}
	return result
}

func printMissingToolsWarning() {
	missing := []string{}
	for _, t := range []string{"semgrep", "trivy", "gitleaks", "bandit"} {
		if _, err := exec.LookPath(t); err != nil {
			missing = append(missing, t)
		}
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "\033[38;2;231;196;39m⚠\033[0m  \033[38;2;132;148;149mMissing: %s\033[0m\n", strings.Join(missing, ", "))
		fmt.Fprintf(os.Stderr, "   \033[38;2;132;148;149mInstall Docker for full SAST scanning, or:\033[0m \033[38;2;0;245;255mbrew install %s\033[0m\n\n", strings.Join(missing, " "))
	}
}

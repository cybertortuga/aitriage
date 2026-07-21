package external

import (
	"bytes"
	"context"
	"os/exec"
)

// RunResult holds the result of running an external tool
type RunResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// RunTool runs an external CLI tool and returns its output.
// Does not panic on non-zero exit codes — just returns it in ExitCode.
func RunTool(ctx context.Context, name string, args ...string) (RunResult, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	// CommandContext reports a killed process as *exec.ExitError. Treating that
	// like an ordinary scanner exit code would turn a timeout/cancellation into
	// a successful zero-finding scan and violate the full-audit fail-closed
	// contract. Preserve the captured output for diagnostics and propagate the
	// context error so the orchestrator records timed_out/failed explicitly.
	if ctxErr := ctx.Err(); ctxErr != nil {
		return RunResult{
			Stdout:   outBuf.String(),
			Stderr:   errBuf.String(),
			ExitCode: -1,
		}, ctxErr
	}
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
		_ = err // non-zero exit code is not an error, the tool works that way
	} else if err != nil {
		return RunResult{}, err
	}
	return RunResult{
		Stdout:   outBuf.String(),
		Stderr:   errBuf.String(),
		ExitCode: exitCode,
	}, nil
}

// IsInstalled checks whether a tool is available in PATH
func IsInstalled(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

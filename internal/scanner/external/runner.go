package external

import (
	"bytes"
	"context"
	"os/exec"
)

// RunResult содержит результат запуска внешнего инструмента
type RunResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// RunTool запускает внешний CLI-инструмент и возвращает его вывод.
// Не паникует при ненулевом exit code — просто возвращает его в ExitCode.
func RunTool(ctx context.Context, name string, args ...string) (RunResult, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
		_ = err // ненулевой exit code — не ошибка, инструмент так работает
	} else if err != nil {
		return RunResult{}, err
	}
	return RunResult{
		Stdout:   outBuf.String(),
		Stderr:   errBuf.String(),
		ExitCode: exitCode,
	}, nil
}

// IsInstalled проверяет доступность инструмента в PATH
func IsInstalled(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

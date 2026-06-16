package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var installMCPCmd = &cobra.Command{
	Use:   "install-mcp",
	Short: "Install AITriage as MCP server in Claude Desktop",
	Long:  "Automatically adds AITriage to Claude Desktop MCP configuration.",
	RunE:  runInstallMCP,
}

func init() {
	rootCmd.AddCommand(installMCPCmd)
}

func runInstallMCP(cmd *cobra.Command, args []string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine binary path: %w", err)
	}

	configPath, err := claudeDesktopConfigPath()
	if err != nil {
		// Если не нашли — вывести инструкцию для ручной установки
		printManualInstall(binaryPath)
		return nil
	}

	// Читаем существующий конфиг или создаём новый
	data, err := os.ReadFile(configPath)
	var config map[string]interface{}
	if err != nil {
		config = map[string]interface{}{}
	} else {
		_ = json.Unmarshal(data, &config)
	}

	// Добавляем AITriage
	if config["mcpServers"] == nil {
		config["mcpServers"] = map[string]interface{}{}
	}
	mcpServers := config["mcpServers"].(map[string]interface{})
	mcpServers["aitriage"] = map[string]interface{}{
		"command": binaryPath,
		"args":    []string{"serve"},
	}

	// Сохраняем
	out, _ := json.MarshalIndent(config, "", "  ")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(configPath, out, 0644); err != nil {
		return err
	}

	fmt.Printf("✅ AITriage added to Claude Desktop MCP config:\n%s\n\n", configPath)
	fmt.Println("Restart Claude Desktop to apply changes.")
	return nil
}

func claudeDesktopConfigPath() (string, error) {
	var configDir string
	switch runtime.GOOS {
	case "darwin":
		configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Claude")
	case "linux":
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "Claude")
	case "windows":
		configDir = filepath.Join(os.Getenv("APPDATA"), "Claude")
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return filepath.Join(configDir, "claude_desktop_config.json"), nil
}

func printManualInstall(binaryPath string) {
	snippet := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"aitriage": map[string]interface{}{
				"command": binaryPath,
				"args":    []string{"serve"},
			},
		},
	}
	data, _ := json.MarshalIndent(snippet, "", "  ")
	fmt.Println("Add this to your Claude Desktop config manually:")
	fmt.Println(string(data))
}

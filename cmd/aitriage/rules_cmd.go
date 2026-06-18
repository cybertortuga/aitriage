package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage rule packs — install, list, and remove custom rule sets",
	Long: `Rule pack management for extending AITriage's detection capabilities.

  aitriage rules list                          → List installed rule packs
  aitriage rules install owasp-api-2025        → Install from registry
  aitriage rules install ./my-rules/           → Install from local directory
  aitriage rules remove owasp-api-2025         → Remove a rule pack`,
}

var rulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed and available rule packs",
	RunE:  runRulesList,
}

var rulesInstallCmd = &cobra.Command{
	Use:   "install [pack-name-or-path]",
	Short: "Install a rule pack from registry or local path",
	Args:  cobra.ExactArgs(1),
	RunE:  runRulesInstall,
}

var rulesRemoveCmd = &cobra.Command{
	Use:   "remove [pack-name]",
	Short: "Remove an installed rule pack",
	Args:  cobra.ExactArgs(1),
	RunE:  runRulesRemove,
}

var rulesInfoCmd = &cobra.Command{
	Use:   "info [pack-name]",
	Short: "Show details about a rule pack",
	Args:  cobra.ExactArgs(1),
	RunE:  runRulesInfo,
}

func init() {
	rootCmd.AddCommand(rulesCmd)
	rulesCmd.AddCommand(rulesListCmd)
	rulesCmd.AddCommand(rulesInstallCmd)
	rulesCmd.AddCommand(rulesRemoveCmd)
	rulesCmd.AddCommand(rulesInfoCmd)
}

// PackManifest describes a rule pack
type PackManifest struct {
	Name        string   `json:"name" yaml:"name"`
	Version     string   `json:"version" yaml:"version"`
	Description string   `json:"description" yaml:"description"`
	Author      string   `json:"author" yaml:"author"`
	RuleCount   int      `json:"rule_count" yaml:"rule_count"`
	Tags        []string `json:"tags" yaml:"tags"`
	Source      string   `json:"source,omitempty" yaml:"source,omitempty"` // GitHub URL or local path
}

// Registry of known rule packs
var registryPacks = map[string]struct {
	Description string
	URL         string
	RuleCount   int
}{
	"owasp-api-2025": {
		Description: "OWASP API Security Top 10 (2025 Edition)",
		URL:         "https://github.com/cybertortuga/rules-owasp-api/releases/latest/download/rules.tar.gz",
		RuleCount:   25,
	},
	"owasp-llm-2025": {
		Description: "OWASP LLM Top 10 — AI/ML Security Rules",
		URL:         "https://github.com/cybertortuga/rules-owasp-llm/releases/latest/download/rules.tar.gz",
		RuleCount:   18,
	},
	"cloud-aws": {
		Description: "AWS Infrastructure Security (IAM, S3, Lambda, etc.)",
		URL:         "https://github.com/cybertortuga/rules-cloud-aws/releases/latest/download/rules.tar.gz",
		RuleCount:   30,
	},
	"cloud-gcp": {
		Description: "GCP Infrastructure Security (IAM, GCS, Cloud Run, etc.)",
		URL:         "https://github.com/cybertortuga/rules-cloud-gcp/releases/latest/download/rules.tar.gz",
		RuleCount:   22,
	},
	"kubernetes": {
		Description: "Kubernetes Security — manifests, RBAC, network policies",
		URL:         "https://github.com/cybertortuga/rules-kubernetes/releases/latest/download/rules.tar.gz",
		RuleCount:   28,
	},
	"pci-dss": {
		Description: "PCI DSS v4.0 Compliance Rules",
		URL:         "https://github.com/cybertortuga/rules-pci-dss/releases/latest/download/rules.tar.gz",
		RuleCount:   35,
	},
	"hipaa": {
		Description: "HIPAA Technical Safeguards",
		URL:         "https://github.com/cybertortuga/rules-hipaa/releases/latest/download/rules.tar.gz",
		RuleCount:   20,
	},
}

func getPacksDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aitriage", "packs")
}

func runRulesList(cmd *cobra.Command, args []string) error {
	cyan := "\033[38;2;0;245;255m"
	green := "\033[38;2;46;204;113m"
	dim := "\033[38;2;120;120;140m"
	bold := "\033[1m"
	reset := "\033[0m"

	fmt.Fprintf(os.Stderr, "\n%s%s  ⦿ AITriage Rule Packs%s\n\n", cyan, bold, reset)

	// List installed packs
	packsDir := getPacksDir()
	installed := make(map[string]PackManifest)

	entries, _ := os.ReadDir(packsDir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		manifestPath := filepath.Join(packsDir, e.Name(), "manifest.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		var m PackManifest
		if json.Unmarshal(data, &m) == nil {
			installed[e.Name()] = m
		}
	}

	if len(installed) > 0 {
		fmt.Fprintf(os.Stderr, "%s  Installed:%s\n", bold, reset)
		for name, m := range installed {
			fmt.Fprintf(os.Stderr, "  %s✓ %s%s v%s — %s (%d rules)%s\n",
				green, bold, name, m.Version, m.Description, m.RuleCount, reset)
		}
		fmt.Fprintln(os.Stderr)
	}

	fmt.Fprintf(os.Stderr, "%s  Available from Registry:%s\n", bold, reset)
	for name, pack := range registryPacks {
		status := "  "
		if _, ok := installed[name]; ok {
			status = green + "✓ " + reset
		} else {
			status = dim + "  " + reset
		}
		fmt.Fprintf(os.Stderr, "  %s%s%s%s — %s (%d rules)%s\n",
			status, cyan, name, dim, pack.Description, pack.RuleCount, reset)
	}

	fmt.Fprintf(os.Stderr, "\n%s  Install: aitriage rules install <pack-name>%s\n\n", dim, reset)
	return nil
}

func runRulesInstall(cmd *cobra.Command, args []string) error {
	packName := args[0]

	cyan := "\033[38;2;0;245;255m"
	green := "\033[38;2;46;204;113m"
	yellow := "\033[38;2;255;214;0m"
	dim := "\033[38;2;120;120;140m"
	bold := "\033[1m"
	reset := "\033[0m"

	packsDir := getPacksDir()
	packDir := filepath.Join(packsDir, packName)

	// Check if it's a local path
	if info, err := os.Stat(packName); err == nil && info.IsDir() {
		return installFromLocal(packName, packsDir, cyan, green, dim, bold, reset)
	}

	// Check registry
	regPack, ok := registryPacks[packName]
	if !ok {
		// Try as GitHub URL
		if strings.HasPrefix(packName, "github.com/") || strings.HasPrefix(packName, "https://") {
			return installFromURL(packName, packName, packsDir, cyan, green, dim, bold, reset)
		}
		fmt.Fprintf(os.Stderr, "%s  ⚠ Pack '%s' not found in registry.%s\n", yellow, packName, reset)
		fmt.Fprintf(os.Stderr, "%s  Available packs: aitriage rules list%s\n\n", dim, reset)
		return nil
	}

	fmt.Fprintf(os.Stderr, "\n%s%s  ⦿ Installing: %s%s\n", cyan, bold, packName, reset)
	fmt.Fprintf(os.Stderr, "%s  %s%s\n", dim, regPack.Description, reset)

	// Download
	fmt.Fprintf(os.Stderr, "%s  Downloading...%s\n", dim, reset)
	resp, err := http.Get(regPack.URL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: HTTP %d (pack may not be published yet)", resp.StatusCode)
	}

	// Extract tar.gz
	if err := os.MkdirAll(packDir, 0755); err != nil {
		return err
	}

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to decompress: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	fileCount := 0
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read error: %w", err)
		}

		target := filepath.Join(packDir, header.Name)
		// Security: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(packDir)) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			_ = os.MkdirAll(target, 0755)
		case tar.TypeReg:
			_ = os.MkdirAll(filepath.Dir(target), 0755)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				continue
			}
			_, _ = io.Copy(f, tr)
			f.Close()
			fileCount++
		}
	}

	// Create manifest if not present
	manifestPath := filepath.Join(packDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		manifest := PackManifest{
			Name:        packName,
			Version:     "latest",
			Description: regPack.Description,
			RuleCount:   regPack.RuleCount,
			Source:      regPack.URL,
		}
		data, _ := json.MarshalIndent(manifest, "", "  ")
		_ = os.WriteFile(manifestPath, data, 0644)
	}

	fmt.Fprintf(os.Stderr, "\n%s%s  ✅ Installed: %s (%d files)%s\n", green, bold, packName, fileCount, reset)
	fmt.Fprintf(os.Stderr, "%s  Location: %s%s\n", dim, packDir, reset)
	fmt.Fprintf(os.Stderr, "%s  Rules will be loaded automatically on next scan.%s\n\n", dim, reset)

	return nil
}

func installFromLocal(localPath, packsDir, cyan, green, dim, bold, reset string) error {
	absPath, _ := filepath.Abs(localPath)
	packName := filepath.Base(absPath)
	packDir := filepath.Join(packsDir, packName)

	fmt.Fprintf(os.Stderr, "\n%s%s  ⦿ Installing from local: %s%s\n", cyan, bold, absPath, reset)

	if err := os.MkdirAll(packDir, 0755); err != nil {
		return err
	}

	// Copy YAML files
	fileCount := 0
	_ = filepath.WalkDir(absPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}
		rel, _ := filepath.Rel(absPath, path)
		dst := filepath.Join(packDir, rel)
		_ = os.MkdirAll(filepath.Dir(dst), 0755)

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		_ = os.WriteFile(dst, data, 0644)
		fileCount++
		return nil
	})

	// Create manifest
	manifest := PackManifest{
		Name:      packName,
		Version:   "local",
		Source:    absPath,
		RuleCount: fileCount,
	}
	data, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile(filepath.Join(packDir, "manifest.json"), data, 0644)

	fmt.Fprintf(os.Stderr, "%s%s  ✅ Installed: %s (%d rule files)%s\n\n", green, bold, packName, fileCount, reset)
	return nil
}

func installFromURL(name, url, packsDir, cyan, green, dim, bold, reset string) error {
	fmt.Fprintf(os.Stderr, "%s  Installing from URL: %s%s\n", dim, url, reset)
	fmt.Fprintf(os.Stderr, "%s  ⚠ URL-based installation coming in v1.1%s\n\n", dim, reset)
	return nil
}

func runRulesRemove(cmd *cobra.Command, args []string) error {
	packName := args[0]
	packDir := filepath.Join(getPacksDir(), packName)

	if _, err := os.Stat(packDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Pack '%s' is not installed.\n", packName)
		return nil
	}

	if err := os.RemoveAll(packDir); err != nil {
		return fmt.Errorf("failed to remove pack: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✅ Removed rule pack: %s\n", packName)
	return nil
}

func runRulesInfo(cmd *cobra.Command, args []string) error {
	packName := args[0]

	// Check installed
	manifestPath := filepath.Join(getPacksDir(), packName, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err == nil {
		var m PackManifest
		if json.Unmarshal(data, &m) == nil {
			cyan := "\033[38;2;0;245;255m"
			dim := "\033[38;2;120;120;140m"
			bold := "\033[1m"
			reset := "\033[0m"

			fmt.Fprintf(os.Stderr, "\n%s%s  %s%s v%s\n", cyan, bold, m.Name, reset, m.Version)
			if m.Description != "" {
				fmt.Fprintf(os.Stderr, "%s  %s%s\n", dim, m.Description, reset)
			}
			fmt.Fprintf(os.Stderr, "%s  Rules: %d%s\n", dim, m.RuleCount, reset)
			if m.Author != "" {
				fmt.Fprintf(os.Stderr, "%s  Author: %s%s\n", dim, m.Author, reset)
			}
			if m.Source != "" {
				fmt.Fprintf(os.Stderr, "%s  Source: %s%s\n", dim, m.Source, reset)
			}
			if len(m.Tags) > 0 {
				fmt.Fprintf(os.Stderr, "%s  Tags: %s%s\n", dim, strings.Join(m.Tags, ", "), reset)
			}
			fmt.Fprintln(os.Stderr)
			return nil
		}
	}

	// Check registry
	if pack, ok := registryPacks[packName]; ok {
		fmt.Fprintf(os.Stderr, "\n  %s (not installed)\n  %s\n  Rules: %d\n\n",
			packName, pack.Description, pack.RuleCount)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Pack '%s' not found.\n", packName)
	return nil
}

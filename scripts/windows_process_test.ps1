<#
.SYNOPSIS
  Native Windows process tests for a built aitriage.exe.

.DESCRIPTION
  Runs the real Windows CLI as a child process. It verifies project-local Codex
  and Claude Code fallback configuration, idempotent uninstall, path handling,
  and setup status classification through a controlled docker.exe shim.

  This is an automated process test. It is not a live Docker Desktop, Codex, or
  Claude Code subscription E2E.
#>

#Requires -Version 5.1
[CmdletBinding()]
param(
  [Parameter(Mandatory = $true)]
  [string]$AitriageBin,
  [string]$GoBin = 'go'
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
if (Get-Variable -Name PSNativeCommandUseErrorActionPreference -ErrorAction SilentlyContinue) {
  $PSNativeCommandUseErrorActionPreference = $false
}

$script:Passes = 0
$script:Failures = 0
function Check([bool]$Condition, [string]$Name) {
  if ($Condition) { $script:Passes++; Write-Host "  PASS  $Name" -ForegroundColor Green }
  else { $script:Failures++; Write-Host "  FAIL  $Name" -ForegroundColor Red }
}

function Invoke-CLI([string[]]$Arguments) {
  $output = & $script:Binary @Arguments 2>&1 | Out-String
  return [pscustomobject]@{ Exit = $LASTEXITCODE; Out = $output }
}

function Count-ExactLine([string]$Text, [string]$Line) {
  return @($Text -split "`r?`n" | Where-Object { $_ -eq $Line }).Count
}

$script:Binary = (Resolve-Path -LiteralPath $AitriageBin).Path
$go = (Get-Command $GoBin -ErrorAction Stop).Source
$work = Join-Path ([IO.Path]::GetTempPath()) ('aitriage-windows-process-' + [IO.Path]::GetRandomFileName())
$project = Join-Path $work 'Project with spaces - данные'
$shim = Join-Path $work 'docker-shim'
$savedPath = $env:PATH
$savedProgramFiles = $env:ProgramFiles
$savedProgramW6432 = $env:ProgramW6432
$savedProgramFilesX86 = ${env:ProgramFiles(x86)}

try {
  New-Item -ItemType Directory -Force -Path $project, $shim | Out-Null

  $versionResult = Invoke-CLI @('version')
  Check ($versionResult.Exit -eq 0 -and $versionResult.Out -match '^AITriage ') 'real aitriage.exe version runs as a process'
  $hostVersion = (($versionResult.Out.Trim() -split '\s+') | Select-Object -Last 1).TrimStart('v')

  # Build a deterministic docker.exe process shim before PATH isolation.
  $stubSource = Join-Path $work 'docker_stub.go'
  @'
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	mode := os.Getenv("AITRIAGE_FAKE_DOCKER_MODE")
	args := os.Args[1:]
	if len(args) == 0 {
		os.Exit(2)
	}
	switch args[0] {
	case "info":
		if mode == "stopped" {
			fmt.Fprintln(os.Stderr, "Docker Desktop is not running")
			os.Exit(1)
		}
		fmt.Println("27.0.0")
	case "image":
		if mode == "image-missing" {
			os.Exit(1)
		}
		if strings.Contains(strings.Join(args, " "), "RepoDigests") {
			fmt.Println("ghcr.io/cybertortuga/aitriage@sha256:test")
		}
	case "run":
		version := os.Getenv("AITRIAGE_FAKE_BUNDLE_VERSION")
		fmt.Printf("aitriage\tok\tAITriage %s\n", version)
		fmt.Println("semgrep\tok\tsemgrep 1.0.0")
		fmt.Println("trivy\tok\tVersion: 1.0.0")
		fmt.Println("gitleaks\tok\tgitleaks 1.0.0")
		fmt.Println("bandit\tok\tbandit 1.0.0")
	default:
		os.Exit(2)
	}
}
'@ | Set-Content -LiteralPath $stubSource -Encoding ASCII
  & $go build -o (Join-Path $shim 'docker.exe') $stubSource
  if ($LASTEXITCODE -ne 0) { throw 'failed to build docker.exe process shim' }

  # Keep Claude absent from PATH so the real CLI exercises its documented
  # project-local .mcp.json fallback instead of touching a user account.
  $env:PATH = $shim

  Set-Content -LiteralPath (Join-Path $project '.gitignore') -Value "/foreign/`n" -Encoding ASCII
  New-Item -ItemType Directory -Force -Path (Join-Path $project '.codex') | Out-Null
  Set-Content -LiteralPath (Join-Path $project '.codex\config.toml') -Value "[mcp_servers.foreign]`ncommand = `"foreign`"`n" -Encoding ASCII
  Set-Content -LiteralPath (Join-Path $project 'AGENTS.md') -Value "# Existing Codex instructions`n" -Encoding ASCII
  Set-Content -LiteralPath (Join-Path $project '.mcp.json') -Value '{"mcpServers":{"foreign":{"command":"foreign"}},"keep":true}' -Encoding ASCII
  Set-Content -LiteralPath (Join-Path $project 'CLAUDE.md') -Value "# Existing Claude instructions`n" -Encoding ASCII

  $codexInstall = Invoke-CLI @('install-codex', $project)
  Check ($codexInstall.Exit -eq 0) 'real CLI installs Codex MCP configuration'
  $codexConfig = Get-Content -Raw -LiteralPath (Join-Path $project '.codex\config.toml')
  Check ($codexConfig -match '\[mcp_servers\.foreign\]' -and $codexConfig -match '\[mcp_servers\.aitriage\]') 'Codex install preserves foreign MCP server'
  $tomlBinary = $script:Binary.Replace('\', '\\').Replace('"', '\"')
  $tomlProject = $project.Replace('\', '\\').Replace('"', '\"')
  Check ($codexConfig.Contains($tomlBinary) -and $codexConfig.Contains($tomlProject)) 'Codex config contains the real executable and project root'
  $codexOnce = $codexConfig
  $codexAgain = Invoke-CLI @('install-codex', $project)
  $codexTwice = Get-Content -Raw -LiteralPath (Join-Path $project '.codex\config.toml')
  Check ($codexAgain.Exit -eq 0 -and $codexOnce -eq $codexTwice) 'Codex install is byte-idempotent'

  $claudeInstall = Invoke-CLI @('install-claude-code', $project)
  Check ($claudeInstall.Exit -eq 0 -and $claudeInstall.Out -match 'Pending approval') 'real CLI uses the honest Claude .mcp.json fallback'
  $claudeConfig = Get-Content -Raw -LiteralPath (Join-Path $project '.mcp.json') | ConvertFrom-Json
  Check ($null -ne $claudeConfig.mcpServers.foreign -and $null -ne $claudeConfig.mcpServers.aitriage) 'Claude install preserves foreign MCP server'
  Check ($claudeConfig.mcpServers.aitriage.command -eq $script:Binary) 'Claude config contains the real executable'
  $claudeOnce = Get-Content -Raw -LiteralPath (Join-Path $project '.mcp.json')
  $claudeAgain = Invoke-CLI @('install-claude-code', $project)
  $claudeTwice = Get-Content -Raw -LiteralPath (Join-Path $project '.mcp.json')
  Check ($claudeAgain.Exit -eq 0 -and $claudeOnce -eq $claudeTwice) 'Claude fallback install is byte-idempotent'

  $ignore = Get-Content -Raw -LiteralPath (Join-Path $project '.gitignore')
  Check ($ignore -match '/foreign/' -and (Count-ExactLine $ignore '/aitriage-reports/') -eq 1) 'connectors preserve .gitignore and add reports exactly once'

  $codexRemove = Invoke-CLI @('install-codex', '--uninstall', $project)
  $claudeRemove = Invoke-CLI @('install-claude-code', '--uninstall', $project)
  $codexAfter = Get-Content -Raw -LiteralPath (Join-Path $project '.codex\config.toml')
  $claudeAfter = Get-Content -Raw -LiteralPath (Join-Path $project '.mcp.json') | ConvertFrom-Json
  $agentsAfter = Get-Content -Raw -LiteralPath (Join-Path $project 'AGENTS.md')
  $claudeMdAfter = Get-Content -Raw -LiteralPath (Join-Path $project 'CLAUDE.md')
  Check ($codexRemove.Exit -eq 0 -and $codexAfter -match '\[mcp_servers\.foreign\]' -and $codexAfter -notmatch 'mcp_servers\.aitriage') 'Codex uninstall removes only AITriage'
  $claudeHasForeign = $null -ne $claudeAfter.mcpServers.PSObject.Properties['foreign']
  $claudeHasAITriage = $null -ne $claudeAfter.mcpServers.PSObject.Properties['aitriage']
  Check ($claudeRemove.Exit -eq 0 -and $claudeHasForeign -and -not $claudeHasAITriage) 'Claude uninstall removes only AITriage'
  Check ($agentsAfter -match 'Existing Codex instructions' -and $agentsAfter -notmatch 'AITRIAGE:BEGIN') 'Codex uninstall preserves project instructions'
  Check ($claudeMdAfter -match 'Existing Claude instructions' -and $claudeMdAfter -notmatch 'AITRIAGE:BEGIN') 'Claude uninstall preserves project instructions'

  # Real process status classification with a controlled docker.exe.
  $emptyPrograms = Join-Path $work 'empty-program-files'
  New-Item -ItemType Directory -Force -Path $emptyPrograms | Out-Null
  $env:PATH = Join-Path $work 'empty-path'
  New-Item -ItemType Directory -Force -Path $env:PATH | Out-Null
  $env:ProgramFiles = $emptyPrograms
  $env:ProgramW6432 = $emptyPrograms
  ${env:ProgramFiles(x86)} = $emptyPrograms
  $missing = Invoke-CLI @('setup', '--status', '--json')
  $missingJSON = $missing.Out | ConvertFrom-Json
  Check ($missing.Exit -ne 0 -and $missingJSON.status -eq 'action_required' -and $missingJSON.code -eq 'docker_not_installed') 'setup status reports Docker not installed'

  $env:PATH = $shim
  $env:AITRIAGE_FAKE_DOCKER_MODE = 'stopped'
  $stopped = Invoke-CLI @('setup', '--status', '--json')
  $stoppedJSON = $stopped.Out | ConvertFrom-Json
  Check ($stopped.Exit -ne 0 -and $stoppedJSON.status -eq 'action_required' -and $stoppedJSON.code -eq 'docker_not_running') 'setup status distinguishes Docker stopped'

  $env:AITRIAGE_FAKE_DOCKER_MODE = 'image-missing'
  $imageMissing = Invoke-CLI @('setup', '--status', '--json')
  $imageMissingJSON = $imageMissing.Out | ConvertFrom-Json
  Check ($imageMissing.Exit -ne 0 -and $imageMissingJSON.status -eq 'action_required' -and $imageMissingJSON.code -eq 'image_missing') 'setup status distinguishes missing scanner image'

  $env:AITRIAGE_FAKE_DOCKER_MODE = 'healthy'
  $env:AITRIAGE_FAKE_BUNDLE_VERSION = $hostVersion
  $healthy = Invoke-CLI @('setup', '--status', '--json')
  $healthyJSON = $healthy.Out | ConvertFrom-Json
  Check ($healthy.Exit -eq 0 -and $healthyJSON.status -eq 'ok') 'setup status accepts a complete scanner bundle'
  Check (@($healthyJSON.bundle | Where-Object { $_.ok }).Count -eq 5) 'setup status reports all five bundled tools healthy'
} finally {
  $env:PATH = $savedPath
  $env:ProgramFiles = $savedProgramFiles
  $env:ProgramW6432 = $savedProgramW6432
  ${env:ProgramFiles(x86)} = $savedProgramFilesX86
  Remove-Item Env:AITRIAGE_FAKE_DOCKER_MODE -ErrorAction SilentlyContinue
  Remove-Item Env:AITRIAGE_FAKE_BUNDLE_VERSION -ErrorAction SilentlyContinue
  Remove-Item -Recurse -Force -LiteralPath $work -ErrorAction SilentlyContinue
}

Write-Host "windows_process_test: $script:Passes passed, $script:Failures failed"
if ($script:Failures -gt 0) { exit 1 } else { exit 0 }

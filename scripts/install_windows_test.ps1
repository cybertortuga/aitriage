<#
.SYNOPSIS
  Automated (process-level) tests for scripts/install.ps1.

.DESCRIPTION
  Hermetic, self-contained installer tests. They do NOT depend on the published
  latest release or on the CI runner's real user profile:

    * A tiny version-stamped stub `aitriage.exe` is compiled with `go build`
      (CGO disabled) to stand in for the real CLI. It answers `version` and
      `setup`. This is an installer-mechanics test — it is explicitly NOT a live
      AITriage/Docker/MCP test.
    * A fake release (ZIP + checksums.txt) is served from a local directory via
      the installer's -ReleaseBaseUrl override.
    * User PATH reads/writes are redirected to a file via AITRIAGE_PATH_STORE, so
      the runner's real PATH/registry is never touched.
    * The child TEMP is redirected to an isolated directory to assert cleanup.

  On non-Windows hosts the full lifecycle is intentionally skipped (native
  Windows is x86_64 Windows only); the parser check, the pure helper units and
  the unsupported-platform rejection still run and must pass.
#>

[CmdletBinding()]
param(
  [string]$Installer = (Join-Path $PSScriptRoot 'install.ps1'),
  [string]$GoBin = 'go'
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$script:Failures = 0
$script:Passes = 0
function Check([bool]$Cond, [string]$Name) {
  if ($Cond) { $script:Passes++; Write-Host "  PASS  $Name" -ForegroundColor Green }
  else { $script:Failures++; Write-Host "  FAIL  $Name" -ForegroundColor Red }
}

$OnWindows = $true
$winVar = Get-Variable -Name IsWindows -ErrorAction SilentlyContinue
if ($winVar) { $OnWindows = [bool]$winVar.Value }

$HostExe = [System.Diagnostics.Process]::GetCurrentProcess().Path
Write-Host "Installer test host: $HostExe (PS $($PSVersionTable.PSVersion)); Windows=$OnWindows"

# ── 1. Parser / syntax check (runs on every host; CI runs it under 5.1 and 7) ─
$tokens = $null; $errors = $null
[System.Management.Automation.Language.Parser]::ParseFile(
  (Resolve-Path $Installer).Path, [ref]$tokens, [ref]$errors) | Out-Null
Check ($errors.Count -eq 0) "install.ps1 parses without syntax errors"
if ($errors.Count -gt 0) { $errors | ForEach-Object { Write-Host "    $($_.Message)" -ForegroundColor Red } }

# ── 2. Pure helper units (dot-source; the auto-run is guarded) ────────────────
. $Installer
$secondLoadSucceeded = $true
try { . $Installer } catch { $secondLoadSucceeded = $false }
Check $secondLoadSucceeded "install.ps1 can be loaded repeatedly in the same PowerShell session"
$sumsFile = Join-Path ([IO.Path]::GetTempPath()) ([IO.Path]::GetRandomFileName())
Set-Content -LiteralPath $sumsFile -Value "abc123  aitriage_9.9.9_Windows_x86_64.zip`ndef456  other.zip"
Check ((Get-ChecksumFor 'aitriage_9.9.9_Windows_x86_64.zip' $sumsFile) -eq 'abc123') "Get-ChecksumFor matches exact asset name"
Check ($null -eq (Get-ChecksumFor 'missing.zip' $sumsFile)) "Get-ChecksumFor returns null for a missing asset"
Remove-Item -LiteralPath $sumsFile -Force

$savedVersion = $Version
$Version = '1.2.3unexpected'
$invalidVersionRejected = $false
try { Resolve-Version | Out-Null } catch { $invalidVersionRejected = $_.Exception.Message -match 'invalid release version' }
$Version = $savedVersion
Check $invalidVersionRejected "Resolve-Version rejects a semver with a trailing suffix"

$savedInstallDir = $InstallDir
$InstallDir = 'relative-install-dir'
$resolvedInstallDir = Resolve-InstallDir
Check ([IO.Path]::IsPathRooted($resolvedInstallDir)) "Resolve-InstallDir normalizes a custom directory to an absolute path"
$InstallDir = 'invalid;path'
$semicolonRejected = $false
try { Resolve-InstallDir | Out-Null } catch { $semicolonRejected = $_.Exception.Message -match 'semicolon' }
$InstallDir = $savedInstallDir
Check $semicolonRejected "Resolve-InstallDir rejects a semicolon that would corrupt User PATH"

$env:AITRIAGE_PATH_STORE = Join-Path ([IO.Path]::GetTempPath()) ([IO.Path]::GetRandomFileName())
try {
  Check (Add-ToUserPath 'C:\aitriage\bin') "Add-ToUserPath adds a new entry"
  Check (-not (Add-ToUserPath 'C:\aitriage\bin')) "Add-ToUserPath is idempotent"
  $stored = Get-UserPath
  $count = @($stored -split ';' | Where-Object { $_.TrimEnd('\') -ieq 'C:\aitriage\bin' }).Count
  Check ($count -eq 1) "User PATH contains the entry exactly once"
  Check (Remove-FromUserPath 'C:\aitriage\bin') "Remove-FromUserPath removes the entry"
} finally {
  Remove-Item -LiteralPath $env:AITRIAGE_PATH_STORE -Force -ErrorAction SilentlyContinue
  Remove-Item Env:AITRIAGE_PATH_STORE -ErrorAction SilentlyContinue
}

# ── Child-process invocation helper ──────────────────────────────────────────
function Invoke-Installer([string[]]$InstallerArgs) {
  $out = & $HostExe -NoProfile -File $Installer @InstallerArgs 2>&1 | Out-String
  return [pscustomobject]@{ Exit = $LASTEXITCODE; Out = $out }
}

if (-not $OnWindows) {
  # Real, honest check on non-Windows: the installer must refuse to run.
  $r = Invoke-Installer @('-Version', '9.9.9', '-SkipSetup')
  Check ($r.Exit -ne 0) "non-Windows host is rejected with a non-zero exit"
  Check ($r.Out -match 'Windows') "non-Windows rejection message mentions Windows"
  Write-Host "Full installer lifecycle is Windows-only and was skipped on this host." -ForegroundColor Yellow
  Write-Host "install_windows_test: $script:Passes passed, $script:Failures failed"
  if ($script:Failures -gt 0) { exit 1 } else { exit 0 }
}

# ── 3. Full lifecycle (Windows only) ─────────────────────────────────────────
$work = Join-Path ([IO.Path]::GetTempPath()) ('aitriage-itest-' + [IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $work | Out-Null
$releases = Join-Path $work 'releases'
$isolatedTemp = Join-Path $work 'temp'
New-Item -ItemType Directory -Path $isolatedTemp | Out-Null
$env:AITRIAGE_PATH_STORE = Join-Path $work 'user-path.txt'
$existingPathEntry = 'C:\Existing\bin'
Set-Content -LiteralPath $env:AITRIAGE_PATH_STORE -Value $existingPathEntry -NoNewline
$env:AITRIAGE_RELEASE_BASE_URL = $releases
$env:TEMP = $isolatedTemp
$env:TMP = $isolatedTemp

function New-FakeRelease([string]$Ver) {
  $tagDir = Join-Path $releases "download\v$Ver"
  New-Item -ItemType Directory -Force -Path $tagDir | Out-Null
  $build = Join-Path $work "build-$Ver"
  New-Item -ItemType Directory -Force -Path $build | Out-Null
  $stub = Join-Path $build 'main.go'
  @'
package main

import (
	"fmt"
	"os"
	"strconv"
)

var version = "0.0.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Println("AITriage " + version)
			return
		case "setup":
			code := 0
			if v := os.Getenv("AITRIAGE_STUB_SETUP_EXIT"); v != "" {
				if n, err := strconv.Atoi(v); err == nil {
					code = n
				}
			}
			if code != 0 {
				fmt.Fprintln(os.Stderr, "stub: setup failed")
			}
			os.Exit(code)
		}
	}
	os.Exit(0)
}
'@ | Set-Content -LiteralPath $stub -Encoding ASCII
  $exe = Join-Path $build 'aitriage.exe'
  $env:CGO_ENABLED = '0'
  & $GoBin build -ldflags "-X main.version=$Ver" -o $exe $stub
  if ($LASTEXITCODE -ne 0) { throw "failed to build stub for $Ver" }
  $asset = "aitriage_${Ver}_Windows_x86_64.zip"
  $zip = Join-Path $tagDir $asset
  if (Test-Path $zip) { Remove-Item $zip -Force }
  Compress-Archive -Path $exe -DestinationPath $zip -Force
  $hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $zip).Hash.ToLower()
  Set-Content -LiteralPath (Join-Path $tagDir 'checksums.txt') -Value "$hash  $asset"
  return $tagDir
}

try {
  New-FakeRelease '9.9.9' | Out-Null
  New-FakeRelease '9.9.8' | Out-Null

  $dirA = Join-Path $work 'installA'

  # Fresh install.
  $r = Invoke-Installer @('-Version', '9.9.9', '-InstallDir', $dirA, '-SkipSetup')
  $exeA = Join-Path $dirA 'aitriage.exe'
  Check ($r.Exit -eq 0) "fresh install exits 0"
  Check (Test-Path $exeA) "fresh install places aitriage.exe"
  if (Test-Path $exeA) { Check ((& $exeA version) -eq 'AITriage 9.9.9') "installed binary reports 9.9.9" }
  Check ((Get-Content -Raw $env:AITRIAGE_PATH_STORE) -match [regex]::Escape($dirA)) "install dir added to User PATH"
  Check ((Get-Content -Raw $env:AITRIAGE_PATH_STORE) -match [regex]::Escape($existingPathEntry)) "existing User PATH entry is preserved"

  # Idempotent reinstall: PATH entry stays single.
  $r = Invoke-Installer @('-Version', '9.9.9', '-InstallDir', $dirA, '-SkipSetup')
  Check ($r.Exit -eq 0) "reinstall exits 0"
  $cnt = @((Get-Content -Raw $env:AITRIAGE_PATH_STORE) -split ';' | Where-Object { $_.TrimEnd('\') -ieq $dirA.TrimEnd('\') }).Count
  Check ($cnt -eq 1) "reinstall does not duplicate the PATH entry"

  # Upgrade from a previous version in a fresh dir.
  $dirB = Join-Path $work 'installB'
  Invoke-Installer @('-Version', '9.9.8', '-InstallDir', $dirB, '-SkipSetup') | Out-Null
  $exeB = Join-Path $dirB 'aitriage.exe'
  Check ((& $exeB version) -eq 'AITriage 9.9.8') "pre-upgrade binary reports 9.9.8"
  $r = Invoke-Installer @('-Version', '9.9.9', '-InstallDir', $dirB, '-SkipSetup')
  Check ($r.Exit -eq 0 -and (& $exeB version) -eq 'AITriage 9.9.9') "upgrade replaces the binary with 9.9.9"
  Check (-not (Test-Path -LiteralPath "$exeB.bak")) "successful atomic upgrade removes its rollback backup"
  Check ((Get-ChildItem -LiteralPath $dirB -Filter '.aitriage.exe.new-*' -ErrorAction SilentlyContinue | Measure-Object).Count -eq 0) "successful atomic upgrade removes staged binaries"

  # Checksum mismatch: fail, keep the previously installed binary.
  $tagDir = Join-Path $releases 'download\v9.9.9'
  $goodZip = Join-Path $tagDir 'aitriage_9.9.9_Windows_x86_64.zip'
  $corruptRoot = Join-Path $work 'corrupt'
  New-FakeRelease '9.9.7' | Out-Null   # build machinery reused; ensures go works
  Add-Content -LiteralPath $goodZip -Value 'corruption'
  $r = Invoke-Installer @('-Version', '9.9.9', '-InstallDir', $dirB, '-SkipSetup')
  Check ($r.Exit -ne 0) "checksum mismatch fails"
  Check ($r.Out -match 'checksum') "checksum mismatch message mentions checksum"
  Check ((& $exeB version) -eq 'AITriage 9.9.9') "old binary preserved after checksum mismatch"
  New-FakeRelease '9.9.9' | Out-Null   # restore a valid release

  # Present ZIP but missing exact checksum entry.
  $tag96 = Join-Path $releases 'download\v9.9.6'
  New-Item -ItemType Directory -Force -Path $tag96 | Out-Null
  Copy-Item -LiteralPath (Join-Path $tagDir 'aitriage_9.9.9_Windows_x86_64.zip') -Destination (Join-Path $tag96 'aitriage_9.9.6_Windows_x86_64.zip')
  Set-Content -LiteralPath (Join-Path $tag96 'checksums.txt') -Value 'deadbeef  some-other-asset.zip'
  $r = Invoke-Installer @('-Version', '9.9.6', '-InstallDir', (Join-Path $work 'nosum'), '-SkipSetup')
  Check ($r.Exit -ne 0 -and $r.Out -match 'checksum.*missing') "missing exact checksum entry fails clearly"

  # Missing ZIP but present checksums.
  $tag95 = Join-Path $releases 'download\v9.9.5'
  New-Item -ItemType Directory -Force -Path $tag95 | Out-Null
  Set-Content -LiteralPath (Join-Path $tag95 'checksums.txt') -Value "deadbeef  aitriage_9.9.5_Windows_x86_64.zip"
  $r = Invoke-Installer @('-Version', '9.9.5', '-InstallDir', (Join-Path $work 'nozip'), '-SkipSetup')
  Check ($r.Exit -ne 0) "missing ZIP fails"

  # Unsupported architecture (simulated via PROCESSOR_ARCHITECTURE).
  $savedArch = $env:PROCESSOR_ARCHITECTURE
  $savedArchW6432 = $env:PROCESSOR_ARCHITEW6432
  $env:PROCESSOR_ARCHITECTURE = 'ARM64'
  Remove-Item Env:PROCESSOR_ARCHITEW6432 -ErrorAction SilentlyContinue
  $r = Invoke-Installer @('-Version', '9.9.9', '-InstallDir', (Join-Path $work 'arm'), '-SkipSetup')
  $env:PROCESSOR_ARCHITECTURE = $savedArch
  if ($null -eq $savedArchW6432) { Remove-Item Env:PROCESSOR_ARCHITEW6432 -ErrorAction SilentlyContinue }
  else { $env:PROCESSOR_ARCHITEW6432 = $savedArchW6432 }
  Check ($r.Exit -ne 0 -and $r.Out -match 'architecture') "unsupported architecture fails with a clear message"

  # setup failure propagates and reports next action.
  $env:AITRIAGE_STUB_SETUP_EXIT = '3'
  $r = Invoke-Installer @('-Version', '9.9.9', '-InstallDir', (Join-Path $work 'setupfail'))
  Check ($r.Exit -ne 0) "setup failure yields non-zero exit"
  Check ($r.Out -match 'setup --full') "setup failure shows the retry command"
  Check ($r.Out -notmatch [regex]::Escape($DockerWindowsUrl)) "generic setup failure is not mislabeled as a Docker installation failure"
  # ...but -SkipSetup must never call setup (would otherwise exit 3).
  $r = Invoke-Installer @('-Version', '9.9.9', '-InstallDir', (Join-Path $work 'skip'), '-SkipSetup')
  Check ($r.Exit -eq 0) "-SkipSetup does not invoke setup"
  Remove-Item Env:AITRIAGE_STUB_SETUP_EXIT -ErrorAction SilentlyContinue

  # -RemoveImage is valid only for uninstall.
  $r = Invoke-Installer @('-Version', '9.9.9', '-InstallDir', $dirA, '-SkipSetup', '-RemoveImage')
  Check ($r.Exit -ne 0 -and $r.Out -match 'only together with -Uninstall') "-RemoveImage without -Uninstall is rejected"

  # Plain uninstall removes only its binary/path entry and preserves other PATH.
  $uninstallDir = Join-Path $work 'uninstall'
  Invoke-Installer @('-Version', '9.9.9', '-InstallDir', $uninstallDir, '-SkipSetup') | Out-Null
  $r = Invoke-Installer @('-InstallDir', $uninstallDir, '-Uninstall')
  Check ($r.Exit -eq 0) "plain uninstall exits 0"
  Check (-not (Test-Path -LiteralPath (Join-Path $uninstallDir 'aitriage.exe'))) "plain uninstall removes aitriage.exe"
  $pathAfterUninstall = Get-Content -Raw $env:AITRIAGE_PATH_STORE
  Check ($pathAfterUninstall -notmatch [regex]::Escape($uninstallDir)) "plain uninstall removes only its PATH entry"
  Check ($pathAfterUninstall -match [regex]::Escape($existingPathEntry)) "plain uninstall preserves foreign PATH entries"

  # A failed requested image removal must keep the CLI installed for retry.
  $removeDir = Join-Path $work 'remove-image'
  Invoke-Installer @('-Version', '9.9.9', '-InstallDir', $removeDir, '-SkipSetup') | Out-Null
  $env:AITRIAGE_STUB_SETUP_EXIT = '4'
  $r = Invoke-Installer @('-InstallDir', $removeDir, '-Uninstall', '-RemoveImage')
  Check ($r.Exit -ne 0) "failed scanner-image removal yields non-zero exit"
  Check (Test-Path -LiteralPath (Join-Path $removeDir 'aitriage.exe')) "failed scanner-image removal preserves CLI for retry"
  Remove-Item Env:AITRIAGE_STUB_SETUP_EXIT -ErrorAction SilentlyContinue
  $r = Invoke-Installer @('-InstallDir', $removeDir, '-Uninstall', '-RemoveImage')
  Check ($r.Exit -eq 0 -and -not (Test-Path -LiteralPath (Join-Path $removeDir 'aitriage.exe'))) "successful image removal continues with CLI uninstall"

  # Temp cleanup: no aitriage-install-* dirs left in the isolated TEMP.
  $leftover = Get-ChildItem -Path $isolatedTemp -Directory -Filter 'aitriage-install-*' -ErrorAction SilentlyContinue
  Check (($leftover | Measure-Object).Count -eq 0) "installer cleans up its temporary directory"

} finally {
  Remove-Item Env:AITRIAGE_PATH_STORE -ErrorAction SilentlyContinue
  Remove-Item Env:AITRIAGE_RELEASE_BASE_URL -ErrorAction SilentlyContinue
  Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
  Remove-Item -Recurse -Force -LiteralPath $work -ErrorAction SilentlyContinue
}

Write-Host "install_windows_test: $script:Passes passed, $script:Failures failed"
if ($script:Failures -gt 0) { exit 1 } else { exit 0 }

<#
.SYNOPSIS
  Official AITriage installer for native Windows (preview).

.DESCRIPTION
  Installs the official, version-pinned AITriage CLI (aitriage.exe) for the
  current user without administrator rights, verifies its SHA-256 against the
  published checksums, adds it to the User PATH, and prepares the complete
  Docker-based scanner bundle via `aitriage setup --full`.

  AITriage is a host tool, not a project dependency. Docker Desktop is required
  and is NOT installed by this script.

  Canonical use (PowerShell 5.1 or 7, no administrator):
    irm https://github.com/cybertortuga/aitriage/releases/latest/download/install.ps1 | iex

  Inspect-first (recommended):
    irm https://github.com/cybertortuga/aitriage/releases/latest/download/install.ps1 -OutFile install.ps1
    # review install.ps1
    $installer = [ScriptBlock]::Create((Get-Content -Raw .\install.ps1))
    & $installer

.NOTES
  Test-only overrides (never weaken checksum verification):
    -Repository / $env:AITRIAGE_REPOSITORY   alternate owner/repo
    -ReleaseBaseUrl / $env:AITRIAGE_RELEASE_BASE_URL   http(s) or local/file base
    $env:AITRIAGE_PATH_STORE   redirect User PATH read/write to a file (tests)
#>

#Requires -Version 5.1
[CmdletBinding()]
param(
  [string]$Version,
  [string]$InstallDir,
  [switch]$SkipSetup,
  [switch]$NonInteractive,
  [switch]$Uninstall,
  [switch]$RemoveImage,
  [string]$Repository,
  [string]$ReleaseBaseUrl
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# ── Configuration (with test-only env overrides) ─────────────────────────────
if (-not $Repository)     { $Repository     = if ($env:AITRIAGE_REPOSITORY) { $env:AITRIAGE_REPOSITORY } else { 'cybertortuga/aitriage' } }
if (-not $ReleaseBaseUrl) { $ReleaseBaseUrl = $env:AITRIAGE_RELEASE_BASE_URL }
$DockerWindowsUrl = 'https://docs.docker.com/desktop/setup/install/windows-install/'
$RetryCommand = 'aitriage setup --full'

# ── Output helpers (color only on an interactive, non-redirected console) ─────
$script:UseColor = $true
if ($env:NO_COLOR) { $script:UseColor = $false }
elseif ($NonInteractive) { $script:UseColor = $false }
else {
  try { if ([Console]::IsOutputRedirected) { $script:UseColor = $false } } catch { $script:UseColor = $false }
}

function Write-Line([string]$Text, [string]$Color) {
  if ($script:UseColor -and $Color) { Write-Host $Text -ForegroundColor $Color }
  else { Write-Host $Text }
}
$script:TotalSteps = 5
function Write-Step([int]$N, [string]$Text) { Write-Line ("[{0}/{1}] {2}" -f $N, $script:TotalSteps, $Text) 'Cyan' }
function Write-Ok([string]$Text)            { Write-Line ("  OK  {0}" -f $Text) 'Green' }
function Write-Action([string]$Text)        { Write-Line ("  ACTION  {0}" -f $Text) 'Yellow' }

function Fail([string]$Message, [string]$ActionUrl = '', [string]$Retry = '') {
  $errorObject = [System.Exception]::new($Message)
  if ($ActionUrl) { $errorObject.Data['ActionUrl'] = $ActionUrl }
  if ($Retry) { $errorObject.Data['Retry'] = $Retry }
  throw $errorObject
}

# ── Networking / filesystem primitives ───────────────────────────────────────
function Set-Tls {
  try {
    [Net.ServicePointManager]::SecurityProtocol = `
      [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
  } catch { }
}

# Get-RemoteFile downloads $Url to $Dest. It supports http(s) (real releases) and
# local paths / file:// URIs (hermetic tests). HTTP uses TLS and fails on error.
function Get-RemoteFile([string]$Url, [string]$Dest) {
  if ($Url -match '^(?i)https?://') {
    Set-Tls
    try {
      Invoke-WebRequest -Uri $Url -OutFile $Dest -UseBasicParsing -MaximumRedirection 5
    } catch {
      Fail "download failed: $Url"
    }
  } else {
    $local = $Url -replace '^(?i)file:/{2,3}', ''
    $local = $local -replace '/', '\'
    if (-not (Test-Path -LiteralPath $local)) { Fail "release asset is unavailable: $Url" }
    Copy-Item -LiteralPath $local -Destination $Dest -Force
  }
}

function Get-ChecksumFor([string]$Name, [string]$File) {
  foreach ($line in Get-Content -LiteralPath $File) {
    $parts = $line -split '\s+', 2
    if ($parts.Count -eq 2) {
      $hash = $parts[0].Trim()
      $named = $parts[1].Trim().TrimStart('*')
      if ($named -eq $Name) { return $hash }
    }
  }
  return $null
}

# ── User PATH (redirectable to a file for tests via AITRIAGE_PATH_STORE) ──────
function Get-UserPath {
  if ($env:AITRIAGE_PATH_STORE) {
    if (Test-Path -LiteralPath $env:AITRIAGE_PATH_STORE) { return (Get-Content -Raw -LiteralPath $env:AITRIAGE_PATH_STORE) }
    return ''
  }
  return [Environment]::GetEnvironmentVariable('Path', 'User')
}
function Set-UserPath([string]$Value) {
  if ($env:AITRIAGE_PATH_STORE) { Set-Content -LiteralPath $env:AITRIAGE_PATH_STORE -Value $Value -NoNewline }
  else { [Environment]::SetEnvironmentVariable('Path', $Value, 'User') }
}
function Test-PathContains([string]$PathValue, [string]$Dir) {
  if ($null -eq $PathValue) { return $false }
  foreach ($e in ($PathValue -split ';')) {
    if ($e -ne '' -and ($e.TrimEnd('\') -ieq $Dir.TrimEnd('\'))) { return $true }
  }
  return $false
}
function Add-ToUserPath([string]$Dir) {
  $cur = Get-UserPath
  if ($null -eq $cur) { $cur = '' }
  if (Test-PathContains $cur $Dir) { return $false }
  $new = if ($cur.Trim() -eq '') { $Dir } else { ($cur.TrimEnd(';') + ';' + $Dir) }
  Set-UserPath $new
  return $true
}
function Remove-FromUserPath([string]$Dir) {
  $cur = Get-UserPath
  if ($null -eq $cur -or $cur -eq '') { return $false }
  $kept = @()
  $removed = $false
  foreach ($e in ($cur -split ';')) {
    if ($e -eq '') { continue }
    if ($e.TrimEnd('\') -ieq $Dir.TrimEnd('\')) { $removed = $true; continue }
    $kept += $e
  }
  if ($removed) { Set-UserPath ($kept -join ';') }
  return $removed
}

# ── Version / target resolution ──────────────────────────────────────────────
function Resolve-InstallDir {
  if ($InstallDir) { $dir = $InstallDir }
  else {
    $base = $env:LOCALAPPDATA
    if (-not $base) { Fail 'LOCALAPPDATA is not set; pass -InstallDir explicitly.' }
    $dir = Join-Path $base 'Programs\AITriage\bin'
  }
  if ($dir.Contains(';')) { Fail 'the install directory cannot contain a semicolon because it is added to User PATH' }
  try { return [IO.Path]::GetFullPath($dir) }
  catch { Fail "invalid install directory: '$dir'" }
}

function Resolve-Version {
  if ($Version) {
    $v = $Version.TrimStart('v')
  } else {
    if ($ReleaseBaseUrl) { Fail 'specify -Version when using a custom release base URL' }
    Set-Tls
    $api = "https://api.github.com/repos/$Repository/releases/latest"
    try {
      $rel = Invoke-RestMethod -Uri $api -UseBasicParsing -Headers @{ 'User-Agent' = 'aitriage-install' }
    } catch {
      Fail 'could not resolve the latest official release'
    }
    $v = ([string]$rel.tag_name).TrimStart('v')
  }
  if ($v -notmatch '^\d+\.\d+\.\d+$') { Fail "invalid release version: '$Version'" }
  return $v
}

function Assert-SupportedPlatform {
  $isWin = $true
  $winVar = Get-Variable -Name IsWindows -ErrorAction SilentlyContinue
  if ($winVar) { $isWin = [bool]$winVar.Value }
  if (-not $isWin) { Fail 'install.ps1 supports native Windows only; on macOS/Linux use the official install.sh.' }

  $arch = $env:PROCESSOR_ARCHITECTURE
  if ($env:PROCESSOR_ARCHITEW6432) { $arch = $env:PROCESSOR_ARCHITEW6432 }
  if ($arch -ne 'AMD64') {
    Fail ("unsupported CPU architecture '{0}'. Native Windows is published for x86_64 (AMD64) only. Windows ARM64 is not yet available; use the GitHub release page or WSL2." -f $arch)
  }
}

# ── Uninstall ────────────────────────────────────────────────────────────────
function Invoke-Uninstall([string]$Dir) {
  $target = Join-Path $Dir 'aitriage.exe'
  if ($RemoveImage -and (Test-Path -LiteralPath $target)) {
    Write-Step 1 'Removing the AITriage scanner image...'
    & $target setup --remove-runtime
    if ($LASTEXITCODE -ne 0) {
      Fail 'Could not remove the AITriage scanner image. The CLI was left installed so you can review the error above and retry.' '' ("& `"{0}`" setup --remove-runtime" -f $target)
    }
  }
  if (Test-Path -LiteralPath $target) {
    Remove-Item -LiteralPath $target -Force
    Write-Ok "Removed $target"
  } else {
    Write-Action "No AITriage binary found at $target"
  }
  if ((Test-Path -LiteralPath $Dir) -and -not (Get-ChildItem -LiteralPath $Dir -Force)) {
    Remove-Item -LiteralPath $Dir -Force
  }
  if (Remove-FromUserPath $Dir) { Write-Ok "Removed $Dir from User PATH" }
  Write-Line 'AITriage was uninstalled. Docker Desktop and any scanner image were left in place unless -RemoveImage was given.' 'Green'
}

function Get-AITriageVersion([string]$Binary) {
  try {
    # Keep the native process out of a PowerShell pipeline. In PowerShell 7.6,
    # piping the first native invocation can leave LASTEXITCODE undefined under
    # StrictMode even though the executable ran successfully.
    $output = @(& $Binary version 2>$null)
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) { Fail "binary version check failed with exit code $exitCode" }
    return [string]($output | Select-Object -First 1)
  } catch {
    Fail "could not run the downloaded AITriage binary: $($_.Exception.Message)"
  }
}

# ── Install / update ─────────────────────────────────────────────────────────
function Invoke-Install([string]$Dir) {
  $version = Resolve-Version
  $tag = "v$version"
  $asset = "aitriage_${version}_Windows_x86_64.zip"
  $base = if ($ReleaseBaseUrl) { $ReleaseBaseUrl } else { "https://github.com/$Repository/releases" }
  $downloadUrl = "$base/download/$tag"

  Write-Step 1 "Resolving AITriage $version for Windows/x86_64"

  $tmp = Join-Path ([IO.Path]::GetTempPath()) ('aitriage-install-' + [IO.Path]::GetRandomFileName())
  New-Item -ItemType Directory -Path $tmp | Out-Null
  try {
    Write-Step 2 'Downloading release and verifying SHA-256'
    $zipPath = Join-Path $tmp $asset
    $sumsPath = Join-Path $tmp 'checksums.txt'
    Get-RemoteFile "$downloadUrl/$asset" $zipPath
    Get-RemoteFile "$downloadUrl/checksums.txt" $sumsPath

    $expected = Get-ChecksumFor $asset $sumsPath
    if (-not $expected) { Fail "checksum for $asset is missing" }
    $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $zipPath).Hash
    if ($actual.ToLower() -ne $expected.ToLower()) {
      Fail 'release checksum verification failed; the installed binary was not changed.'
    }
    Write-Ok 'Checksum verified'

    $extractDir = Join-Path $tmp 'extracted'
    Expand-Archive -LiteralPath $zipPath -DestinationPath $extractDir -Force
    $newExe = Join-Path $extractDir 'aitriage.exe'
    if (-not (Test-Path -LiteralPath $newExe)) { Fail 'release archive does not contain aitriage.exe' }

    # Verify the NEW binary's version BEFORE replacing the installed one, using
    # its absolute path (never a possibly-stale PATH entry).
    Write-Step 3 'Verifying downloaded binary'
    $reported = Get-AITriageVersion $newExe
    if ($reported -ne "AITriage $version" -and $reported -ne "AITriage $tag") {
      Fail "release binary reported an unexpected version: '$reported'"
    }
    Write-Ok $reported

    # Stage the verified executable on the destination volume, then replace the
    # old binary atomically. File.Replace keeps a same-volume backup until the
    # installed copy has reported the expected version.
    New-Item -ItemType Directory -Force -Path $Dir | Out-Null
    $target = Join-Path $Dir 'aitriage.exe'
    $backup = "$target.bak"
    $staged = Join-Path $Dir ('.aitriage.exe.new-' + [IO.Path]::GetRandomFileName())
    $hadPrevious = Test-Path -LiteralPath $target
    if (Test-Path -LiteralPath $backup) { Remove-Item -LiteralPath $backup -Force }
    try {
      Copy-Item -LiteralPath $newExe -Destination $staged -Force
      if ($hadPrevious) {
        [IO.File]::Replace($staged, $target, $backup, $true)
      } else {
        [IO.File]::Move($staged, $target)
      }
    } catch {
      Remove-Item -LiteralPath $staged -Force -ErrorAction SilentlyContinue
      Fail "could not install into $target (is a previous aitriage.exe still running?)"
    }

    # Confirm the installed binary (absolute path) reports the expected version.
    $installed = ''
    $installCheckError = $null
    try { $installed = Get-AITriageVersion $target } catch { $installCheckError = $_.Exception }
    if ($installCheckError -or ($installed -ne "AITriage $version" -and $installed -ne "AITriage $tag")) {
      if (Test-Path -LiteralPath $backup) {
        try { [IO.File]::Replace($backup, $target, $null, $true) }
        catch { Copy-Item -LiteralPath $backup -Destination $target -Force }
      } elseif (-not $hadPrevious -and (Test-Path -LiteralPath $target)) {
        Remove-Item -LiteralPath $target -Force -ErrorAction SilentlyContinue
      }
      Remove-Item -LiteralPath $backup -Force -ErrorAction SilentlyContinue
      if ($installCheckError) { Fail "installed binary verification failed; the previous binary was restored: $($installCheckError.Message)" }
      Fail "installed binary reported an unexpected version: '$installed'"
    }
    Remove-Item -LiteralPath $backup -Force -ErrorAction SilentlyContinue

    Write-Step 4 'Updating User PATH'
    if (Add-ToUserPath $Dir) { Write-Ok "Added $Dir to User PATH" } else { Write-Ok 'User PATH already contains the install directory' }
    if (-not (Test-PathContains $env:Path $Dir)) { $env:Path = "$Dir;$env:Path" }
    Write-Line "Installed: $target" 'Green'
    Write-Line $installed 'Green'

    if ($SkipSetup) {
      Write-Line 'AITriage CLI installed. Skipping scanner setup (developer/CI mode).' 'Green'
      return
    }

    Write-Step 5 'Preparing the complete scanner bundle (aitriage setup --full)'
    & $target setup --full
    if ($LASTEXITCODE -ne 0) {
      Fail ("The AITriage CLI is installed at {0}, but scanner setup did not complete. Review the specific setup error above, correct it, then run: {1}" -f $target, $RetryCommand) '' $RetryCommand
    }
    Write-Line 'AITriage is ready.' 'Green'
    Write-Line 'Next: open your project and run  aitriage install-codex .  or  aitriage install-claude-code .' 'Green'
  } finally {
    Remove-Item -Recurse -Force -LiteralPath $tmp -ErrorAction SilentlyContinue
  }
}

# ── Entry point ──────────────────────────────────────────────────────────────
function Invoke-Main {
  if ($RemoveImage -and -not $Uninstall) {
    Fail '-RemoveImage is valid only together with -Uninstall.'
  }
  Assert-SupportedPlatform
  $dir = Resolve-InstallDir
  if ($Uninstall) { Invoke-Uninstall $dir } else { Invoke-Install $dir }
}

# Run only when executed as a script. When dot-sourced (InvocationName '.'),
# only the functions are defined so tests can exercise them in isolation.
if ($MyInvocation.InvocationName -ne '.') {
  try {
    Invoke-Main
  } catch {
    $e = $_.Exception
    $msg = $e.Message
    Write-Line "AITriage install error: $msg" 'Red'
    $actionUrl = [string]$e.Data['ActionUrl']
    $retry = [string]$e.Data['Retry']
    if ($actionUrl) { Write-Line "  See: $actionUrl" 'Yellow' }
    if ($retry) { Write-Line "  Try again: $retry" 'Yellow' }
    if ($VerbosePreference -eq 'Continue') { Write-Line $_.ScriptStackTrace 'DarkGray' }
    exit 1
  }
}

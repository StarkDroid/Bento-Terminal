# Install Bento terminal on Windows.
#
#   irm https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.ps1 | iex
#
# It downloads the release binary for this machine into %LOCALAPPDATA%\bento and
# adds that folder to your user PATH. Open a new Windows Terminal afterwards.
#
# Override the folder with -InstallDir or BENTO_INSTALL_DIR, and the repository
# with BENTO_REPO (default StarkDroid/Bento-Terminal).

[CmdletBinding()]
param(
	[string]$InstallDir
)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

if (-not $InstallDir) {
	$InstallDir = $env:BENTO_INSTALL_DIR
}
if (-not $InstallDir) {
	$InstallDir = Join-Path $env:LOCALAPPDATA 'bento'
}

$repo = $env:BENTO_REPO
if (-not $repo) {
	$repo = 'StarkDroid/Bento-Terminal'
}

$arch = 'amd64'
if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64' -or $env:PROCESSOR_ARCHITEW6432 -eq 'ARM64') {
	$arch = 'arm64'
}

$asset = "bento_windows_$arch.exe"
$url = "https://github.com/$repo/releases/latest/download/$asset"
$target = Join-Path $InstallDir 'bento.exe'

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

Write-Host "Downloading $asset"
try {
	Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile $target
}
catch {
	throw "Could not download $url. Check the release assets and your connection."
}

if (-not (Test-Path $target) -or (Get-Item $target).Length -eq 0) {
	throw "The download did not produce $target."
}

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$parts = @()
if ($userPath) {
	$parts = $userPath -split ';' | Where-Object { $_ }
}
if ($parts -notcontains $InstallDir) {
	$joined = (@($parts) + $InstallDir) -join ';'
	[Environment]::SetEnvironmentVariable('Path', $joined, 'User')
	Write-Host "Added $InstallDir to your user PATH."
}
$env:Path = "$env:Path;$InstallDir"

Write-Host "Installed $target"
Write-Host 'Open a new Windows Terminal, change to a project, and run: bento'

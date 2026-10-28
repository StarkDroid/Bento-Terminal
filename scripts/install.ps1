# Install Bento terminal on Windows.
#
#   irm https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.ps1 | iex
#   .\scripts\install.ps1
#   .\scripts\install.ps1 -FromSource
#   .\scripts\install.ps1 -FromSource -Source C:\src\Bento-Terminal
#
# The default install downloads the release binary for this machine into
# %LOCALAPPDATA%\bento and adds that folder to your user PATH, so a new Windows
# Terminal finds `bento`. Override the folder with -InstallDir or
# BENTO_INSTALL_DIR, the repository with BENTO_REPO or BENTO_REPO_URL, and leave
# the PATH alone with -NoPath.
#
# -FromSource installs a build of the checkout instead of the release: the
# repository this file lives in, or -Source. It needs Go 1.23 or newer on PATH
# and is how a branch is tried before a release exists. Without the flag the
# release is tried first, then the same source build, which is what
# scripts/install.sh does on macOS and Linux.
#
# The binary is unsigned, so SmartScreen may ask once.

[CmdletBinding()]
param(
	[string]$InstallDir,
	[switch]$FromSource,
	[string]$Source,
	[switch]$NoPath
)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

if ($Source) {
	$FromSource = $true
}
if (-not $InstallDir) {
	$InstallDir = $env:BENTO_INSTALL_DIR
}
if (-not $InstallDir) {
	$InstallDir = Join-Path $env:LOCALAPPDATA 'bento'
}
$InstallDir = $InstallDir.TrimEnd('\', '/')

$repo = $env:BENTO_REPO
if (-not $repo) {
	$repo = 'StarkDroid/Bento-Terminal'
}
$repoUrl = $env:BENTO_REPO_URL
if (-not $repoUrl) {
	$repoUrl = "https://github.com/$repo.git"
}

$arch = 'amd64'
if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64' -or $env:PROCESSOR_ARCHITEW6432 -eq 'ARM64') {
	$arch = 'arm64'
}

$asset = "bento_windows_$arch.exe"
$url = "https://github.com/$repo/releases/latest/download/$asset"
$target = Join-Path $InstallDir 'bento.exe'

function Test-PortableExe($path) {
	# A missing asset can answer with an HTML error page; only a PE starts "MZ".
	$stream = [System.IO.File]::OpenRead($path)
	try {
		return ($stream.ReadByte() -eq 0x4D -and $stream.ReadByte() -eq 0x5A)
	}
	finally {
		$stream.Close()
	}
}

function Install-Release {
	Write-Host "Downloading $asset"
	$partial = "$target.partial"
	try {
		Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile $partial
	}
	catch {
		Remove-Item -Force $partial -ErrorAction SilentlyContinue
		return $false
	}
	if (-not (Test-Path $partial) -or (Get-Item $partial).Length -eq 0 -or -not (Test-PortableExe $partial)) {
		Remove-Item -Force $partial -ErrorAction SilentlyContinue
		return $false
	}
	try {
		Move-Item -Force $partial $target
	}
	catch {
		Remove-Item -Force $partial -ErrorAction SilentlyContinue
		throw "Could not replace $target. Close any running bento window and try again."
	}
	return $true
}

function Find-Checkout {
	if (-not $PSScriptRoot) {
		return $null
	}
	$root = Split-Path -Parent $PSScriptRoot
	if ((Test-Path (Join-Path $root 'go.mod')) -and (Test-Path (Join-Path $root 'cmd\bento'))) {
		return $root
	}
	return $null
}

function Install-FromSource {
	$tree = $null
	if ($Source) {
		$tree = (Resolve-Path -Path $Source).Path
		if (-not (Test-Path (Join-Path $tree 'go.mod')) -or -not (Test-Path (Join-Path $tree 'cmd\bento'))) {
			throw "$tree is not a Bento checkout: it has no go.mod and cmd\bento."
		}
	}
	else {
		$tree = Find-Checkout
	}
	$clone = $null
	if (-not $tree) {
		if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
			throw "git is required to clone https://github.com/$repo and build from source."
		}
		$clone = Join-Path ([System.IO.Path]::GetTempPath()) ("bento-" + [guid]::NewGuid().ToString('N'))
		Write-Host "Cloning $repoUrl"
		git clone --depth 1 $repoUrl $clone
		if ($LASTEXITCODE -ne 0) {
			throw "git clone failed."
		}
		$tree = $clone
	}
	try {
		if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
			throw "Building from source needs Go 1.23 or newer on PATH. A published release binary is the install path that does not need a toolchain."
		}
		Write-Host "Building bento from $tree"
		Push-Location $tree
		try {
			go build -o $target ./cmd/bento
			if ($LASTEXITCODE -ne 0) {
				throw "go build failed. Close any running bento window and try again, or fix the errors above."
			}
		}
		finally {
			Pop-Location
		}
	}
	finally {
		if ($clone) {
			Remove-Item -Recurse -Force $clone -ErrorAction SilentlyContinue
		}
	}
}

function Get-UserPath {
	$key = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment')
	if (-not $key) {
		return @{ Value = ''; Kind = 'String' }
	}
	try {
		# Read unexpanded so %USERPROFILE%-style entries survive the round trip.
		$value = $key.GetValue('Path', '', [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
		$kind = 'String'
		try {
			$kind = $key.GetValueKind('Path').ToString()
		}
		catch {
			# No Path value yet.
		}
		return @{ Value = [string]$value; Kind = $kind }
	}
	finally {
		$key.Close()
	}
}

function Send-EnvironmentChange {
	# Explorer caches the environment block, so a Windows Terminal opened from it
	# only sees the new PATH after this broadcast. Best effort: the PATH is
	# already written when the type cannot be compiled.
	$source = @'
using System;
using System.Runtime.InteropServices;
public static class BentoEnvironment {
	[DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
	public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint msg, UIntPtr wParam, string lParam, uint flags, uint timeout, out UIntPtr result);
	public static void Broadcast() {
		UIntPtr result;
		SendMessageTimeout(new IntPtr(0xffff), 0x1A, UIntPtr.Zero, "Environment", 0x2, 5000, out result);
	}
}
'@
	try {
		Add-Type -TypeDefinition $source -ErrorAction Stop
		[BentoEnvironment]::Broadcast()
	}
	catch {
		Write-Host "Restart Windows Terminal or sign out to pick up the new PATH."
	}
}

function Add-UserPath($dir) {
	$current = Get-UserPath
	$parts = @()
	if ($current.Value) {
		$parts = $current.Value -split ';' | Where-Object { $_ }
	}
	if ($parts -contains $dir) {
		Write-Host "$dir is already on your user PATH."
		return
	}
	$joined = (@($parts) + $dir) -join ';'
	$key = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment', $true)
	try {
		$kind = [Microsoft.Win32.RegistryValueKind]::String
		if ($current.Kind -eq 'ExpandString') {
			$kind = [Microsoft.Win32.RegistryValueKind]::ExpandString
		}
		$key.SetValue('Path', $joined, $kind)
	}
	finally {
		$key.Close()
	}
	Write-Host "Added $dir to your user PATH."
	Send-EnvironmentChange
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

if ($FromSource) {
	Install-FromSource
}
elseif (Install-Release) {
	Write-Host "Installed the published release binary."
}
else {
	Write-Host "No release binary named $asset could be downloaded."
	Write-Host "Building from source instead; this needs Go 1.23 or newer."
	Install-FromSource
}

if (-not (Test-Path $target) -or (Get-Item $target).Length -eq 0) {
	throw "The install did not produce $target."
}

& $target --version
if ($LASTEXITCODE -ne 0) {
	throw "$target did not run."
}

Write-Host "Installed $target"
if ($NoPath) {
	Write-Host "PATH was not changed. Run: $target"
}
else {
	Add-UserPath $InstallDir
	$env:Path = "$env:Path;$InstallDir"
	Write-Host 'Open a new Windows Terminal, change to a project, and run: bento'
}

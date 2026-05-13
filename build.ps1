# Build script for nextdns-client
# Cross-platform build using PowerShell
param(
    [string]$Task = "build",
    [string]$BinaryName = "nextdns-client",
    [string]$GoPackage = "nextdns_client",
    [string]$GoFiles = "./cmd",
    [string]$BuildDir = "build",
    [string]$DistDir = "dist"
)

# Ensure we're in the project root
$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Definition

# Helper: Get GOOS/GOARCH
function Get-GOOS {
    return (go env GOOS).Trim()
}

function Get-GOARCH {
    return (go env GOARCH).Trim()
}

# Helper: Get version from git
function Get-Version {
    try {
        $version = git describe --tags --always --dirty 2>$null
        if (-not $version) {
            $version = "v0.0.0-unknown"
        }
        return $version
    } catch {
        return "v0.0.0-unknown"
    }
}

# Helper: Get build date
function Get-BuildDate {
    return ((Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ'))
}

# Helper: Create directory
function Ensure-Dir {
    param([string]$Path)
    if (-not (Test-Path -Path $Path)) {
        New-Item -ItemType Directory -Path $Path | Out-Null
    }
}

# Build function
function Start-Build {
    param(
        [string]$OutputFile,
        [string]$LDFlags,
        [string]$TargetOS,
        [string]$TargetArch
    )
    $output = [System.IO.Path]::GetFullPath($OutputFile)
    Write-Host "  Building for $TargetOS/$TargetArch -> $output" -ForegroundColor Cyan
    
    # Save current env
    $oldOS = $env:GOOS
    $oldArch = $env:GOARCH
    
    try {
        $env:GOOS = $TargetOS
        $env:GOARCH = $TargetArch
        go build -ldflags="$LDFlags" -o $output $GoFiles
    }
    finally {
        # Restore env
        $env:GOOS = $oldOS
        $env:GOARCH = $oldArch
    }
}

# Main logic
$ErrorActionPreference = "Stop"

$HOST_OS = $(Get-GOOS)
$HOST_ARCH = $(Get-GOARCH)
$VERSION = $(Get-Version)
$BUILD_DATE = $(Get-BuildDate)
$GO_LDFLAGS = "-X main.version=$VERSION -X main.buildDate=$BUILD_DATE"

Write-Host "nextdns-client build system"

switch ($Task.ToLower()) {
    "build" {
        Write-Host "Build target: $Task"
        Write-Host "  GOOS: $HOST_OS"
        Write-Host "  GOARCH: $HOST_ARCH"
        Write-Host "  Version: $VERSION"
        Write-Host "  Date: $BUILD_DATE"

        Ensure-Dir "$ProjectRoot/$BuildDir"
        $ext = ""
        if ($HOST_OS -eq "windows") {
            $ext = ".exe"
        }
        Start-Build -OutputFile "$ProjectRoot/$BuildDir/$BinaryName$ext" -LDFlags $GO_LDFLAGS -TargetOS $HOST_OS -TargetArch $HOST_ARCH
    }

    "dist" {
        Ensure-Dir "$ProjectRoot/$DistDir"

        Write-Host "Cross-platform distribution build"

        # Windows x64
        Start-Build -OutputFile "$ProjectRoot/$DistDir/$BinaryName.exe" -LDFlags $GO_LDFLAGS -TargetOS "windows" -TargetArch "amd64"

        # Windows x86
        Start-Build -OutputFile "$ProjectRoot/$DistDir/$BinaryName-32bit.exe" -LDFlags $GO_LDFLAGS -TargetOS "windows" -TargetArch "386"

        # Linux x64
        Start-Build -OutputFile "$ProjectRoot/$DistDir/$BinaryName-linux-amd64" -LDFlags $GO_LDFLAGS -TargetOS "linux" -TargetArch "amd64"

        # Linux x86
        Start-Build -OutputFile "$ProjectRoot/$DistDir/$BinaryName-linux-386" -LDFlags $GO_LDFLAGS -TargetOS "linux" -TargetArch "386"

        # macOS x64
        Start-Build -OutputFile "$ProjectRoot/$DistDir/$BinaryName-darwin-amd64" -LDFlags $GO_LDFLAGS -TargetOS "darwin" -TargetArch "amd64"

        # macOS ARM64
        Start-Build -OutputFile "$ProjectRoot/$DistDir/$BinaryName-darwin-arm64" -LDFlags $GO_LDFLAGS -TargetOS "darwin" -TargetArch "arm64"

        Write-Host "Cross-platform binaries created successfully" -ForegroundColor Green
    }

    "win64" {
        Write-Host "Build target: $Task"
        Ensure-Dir "$ProjectRoot/$DistDir"
        Start-Build -OutputFile "$ProjectRoot/$DistDir/$BinaryName.exe" -LDFlags $GO_LDFLAGS -TargetOS "windows" -TargetArch "amd64"
    }

    "win32" {
        Write-Host "Build target: $Task"
        Ensure-Dir "$ProjectRoot/$DistDir"
        Start-Build -OutputFile "$ProjectRoot/$DistDir/$BinaryName-32bit.exe" -LDFlags $GO_LDFLAGS -TargetOS "windows" -TargetArch "386"
    }

    "linux64" {
        Write-Host "Build target: $Task"
        Ensure-Dir "$ProjectRoot/$DistDir"
        Start-Build -OutputFile "$ProjectRoot/$DistDir/$BinaryName-linux-amd64" -LDFlags $GO_LDFLAGS -TargetOS "linux" -TargetArch "amd64"
    }

    "linux32" {
        Write-Host "Build target: $Task"
        Ensure-Dir "$ProjectRoot/$DistDir"
        Start-Build -OutputFile "$ProjectRoot/$DistDir/$BinaryName-linux-386" -LDFlags $GO_LDFLAGS -TargetOS "linux" -TargetArch "386"
    }

    "mac64" {
        Write-Host "Build target: $Task"
        Ensure-Dir "$ProjectRoot/$DistDir"
        Start-Build -OutputFile "$ProjectRoot/$DistDir/$BinaryName-darwin-amd64" -LDFlags $GO_LDFLAGS -TargetOS "darwin" -TargetArch "amd64"
    }

    "macarm64" {
        Write-Host "Build target: $Task"
        Ensure-Dir "$ProjectRoot/$DistDir"
        Start-Build -OutputFile "$ProjectRoot/$DistDir/$BinaryName-darwin-arm64" -LDFlags $GO_LDFLAGS -TargetOS "darwin" -TargetArch "arm64"
    }

    "cross-platform" {
        Invoke-Expression "pwsh -NoProfile -File `"$script:BuildScriptPath`" -Task dist"
    }

    "clean" {
        Write-Host "Cleaning build artifacts..."
        if (Test-Path "$ProjectRoot/$DistDir") { Remove-Item -Recurse -Force "$ProjectRoot/$DistDir" -ErrorAction SilentlyContinue | Out-Null }
        if (Test-Path "$ProjectRoot/$BuildDir") { Remove-Item -Recurse -Force "$ProjectRoot/$BuildDir" -ErrorAction SilentlyContinue | Out-Null }
        Write-Host "Clean complete" -ForegroundColor Green
    }

    "help" {
        Write-Host "Usage: make build|dist|clean|help"
        Get-Help
    }

    default {
        Write-Host "Unknown task: $Task"
        Write-Host "Run 'make help' for usage"
    }
}

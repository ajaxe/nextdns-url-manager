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
        [string]$LDFlags
    )
    $output = [System.IO.Path]::GetFullPath($OutputFile)
    Write-Host "  Building -> $output" -ForegroundColor Cyan
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    go build -ldflags="$LDFlags" -o $output $GoFiles
    $LASTEXITCODE | Out-Host
}

# Main logic
$ErrorActionPreference = "Stop"

$GOOS = $(Get-GOOS)
$GOARCH = $(Get-GOARCH)
$VERSION = $(Get-Version)
$BUILD_DATE = $(Get-BuildDate)
$GO_LDFLAGS = "-X main.version=$VERSION -X main.buildDate=$BUILD_DATE"

Write-Host "nextdns-client build system"

switch ($Task.ToLower()) {
    "build" {
        Write-Host "Build target: $Task"
        Write-Host "  GOOS: $GOOS"
        Write-Host "  GOARCH: $GOARCH"
        Write-Host "  Version: $VERSION"
        Write-Host "  Date: $BUILD_DATE"

        Ensure-Dir "$ProjectRoot/$BuildDir"
        $ext = ""
        if ($GOOS -eq "windows") {
            $ext = ".exe"
        }
        Start-Build "$ProjectRoot/$BuildDir/$BinaryName$ext" $GO_LDFLAGS
    }

    "dist" {
        Ensure-Dir "$ProjectRoot/$DistDir"

        Write-Host "Cross-platform distribution build"

        # Windows x64
        $env:os = "windows"; $env:arch = "amd64"
        Start-Build "$ProjectRoot/$DistDir/$BinaryName.exe" $GO_LDFLAGS

        # Windows x86
        $env:os = "windows"; $env:arch = "386"
        Start-Build "$ProjectRoot/$DistDir/$BinaryName-32bit.exe" $GO_LDFLAGS

        # Linux x64
        $env:os = "linux"; $env:arch = "amd64"
        Start-Build "$ProjectRoot/$DistDir/$BinaryName-linux-amd64" $GO_LDFLAGS

        # Linux x86
        $env:os = "linux"; $env:arch = "386"
        Start-Build "$ProjectRoot/$DistDir/$BinaryName-linux-386" $GO_LDFLAGS

        # macOS x64
        $env:os = "darwin"; $env:arch = "amd64"
        Start-Build "$ProjectRoot/$DistDir/$BinaryName-darwin-amd64" $GO_LDFLAGS

        # macOS ARM64
        $env:os = "darwin"; $env:arch = "arm64"
        Start-Build "$ProjectRoot/$DistDir/$BinaryName-darwin-arm64" $GO_LDFLAGS

        Write-Host "Cross-platform binaries created successfully" -ForegroundColor Green
    }

    "win64" {
        Write-Host "Build target: $Task"
        Ensure-Dir "$ProjectRoot/$DistDir"
        $env:os = "windows"; $env:arch = "amd64"
        Start-Build "$ProjectRoot/$DistDir/$BinaryName.exe" $GO_LDFLAGS
    }

    "win32" {
        Write-Host "Build target: $Task"
        Ensure-Dir "$ProjectRoot/$DistDir"
        $env:os = "windows"; $env:arch = "386"
        Start-Build "$ProjectRoot/$DistDir/$BinaryName-32bit.exe" $GO_LDFLAGS
    }

    "linux64" {
        Write-Host "Build target: $Task"
        Ensure-Dir "$ProjectRoot/$DistDir"
        $env:os = "linux"; $env:arch = "amd64"
        Start-Build "$ProjectRoot/$DistDir/$BinaryName-linux-amd64" $GO_LDFLAGS
    }

    "linux32" {
        Write-Host "Build target: $Task"
        Ensure-Dir "$ProjectRoot/$DistDir"
        $env:os = "linux"; $env:arch = "386"
        Start-Build "$ProjectRoot/$DistDir/$BinaryName-linux-386" $GO_LDFLAGS
    }

    "mac64" {
        Write-Host "Build target: $Task"
        Ensure-Dir "$ProjectRoot/$DistDir"
        $env:os = "darwin"; $env:arch = "amd64"
        Start-Build "$ProjectRoot/$DistDir/$BinaryName-darwin-amd64" $GO_LDFLAGS
    }

    "macarm64" {
        Write-Host "Build target: $Task"
        Ensure-Dir "$ProjectRoot/$DistDir"
        $env:os = "darwin"; $env:arch = "arm64"
        Start-Build "$ProjectRoot/$DistDir/$BinaryName-darwin-arm64" $GO_LDFLAGS
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

# Pivot CLI Installer for Windows PowerShell
# This script downloads and installs the latest version of Pivot CLI

param(
    [switch]$Help,
    [string]$InstallDir = "$env:USERPROFILE\bin"
)

# Configuration
$Repo = "rhino11/pivot"
$BinaryName = "pivot.exe"

# Colors for output
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
    White = "White"
}

# Helper functions
function Write-Info {
    param([string]$Message)
    Write-Host "ℹ $Message" -ForegroundColor $Colors.Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor $Colors.Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor $Colors.Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor $Colors.Red
    exit 1
}

# Detect architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { Write-Error "Unsupported architecture: $arch" }
    }
}

# Get the latest release version
function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        return $response.tag_name
    }
    catch {
        Write-Error "Failed to get latest version: $($_.Exception.Message)"
    }
}

# Download and install
function Install-Pivot {
    param(
        [string]$Architecture,
        [string]$Version,
        [string]$InstallDirectory
    )
    
    $platform = "windows-$Architecture"
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$BinaryName"
    $downloadUrl = $downloadUrl -replace "\.exe", "-$platform.exe"
    
    # Create install directory if it doesn't exist
    if (!(Test-Path $InstallDirectory)) {
        Write-Info "Creating install directory: $InstallDirectory"
        New-Item -ItemType Directory -Path $InstallDirectory -Force | Out-Null
    }
    
    $installPath = Join-Path $InstallDirectory $BinaryName
    
    Write-Info "Downloading Pivot $Version for $platform..."
    
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $installPath
    }
    catch {
        Write-Error "Failed to download Pivot: $($_.Exception.Message)"
    }
    
    Write-Success "Pivot installed successfully to: $installPath"
    
    # Add to PATH if not already there
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($userPath -notlike "*$InstallDirectory*") {
        Write-Info "Adding $InstallDirectory to your PATH..."
        $newPath = "$userPath;$InstallDirectory"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Success "Added to PATH. You may need to restart your terminal."
    }
}

# Verify installation
function Test-Installation {
    param([string]$InstallDirectory)
    
    $installPath = Join-Path $InstallDirectory $BinaryName
    
    if (Test-Path $installPath) {
        try {
            $version = & $installPath version 2>$null | Select-Object -First 1
            Write-Success "Pivot is installed and ready to use ($version)"
            Write-Info "Run 'pivot --help' to get started"
        }
        catch {
            Write-Warning "Pivot was installed but there was an issue running it"
        }
    }
    else {
        Write-Error "Installation failed - binary not found at $installPath"
    }
}

# Show help
function Show-Help {
    Write-Host "Pivot CLI Installer for Windows" -ForegroundColor $Colors.White
    Write-Host "================================" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Usage: .\install.ps1 [OPTIONS]" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Install the latest version of Pivot CLI" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Options:" -ForegroundColor $Colors.White
    Write-Host "  -Help              Show this help message" -ForegroundColor $Colors.White
    Write-Host "  -InstallDir PATH   Install directory (default: $env:USERPROFILE\bin)" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "This script will:" -ForegroundColor $Colors.White
    Write-Host "  1. Detect your system architecture" -ForegroundColor $Colors.White
    Write-Host "  2. Download the latest Pivot release" -ForegroundColor $Colors.White
    Write-Host "  3. Install it to the specified directory" -ForegroundColor $Colors.White
    Write-Host "  4. Add the directory to your PATH" -ForegroundColor $Colors.White
}

# Main installation process
function Main {
    if ($Help) {
        Show-Help
        return
    }
    
    Write-Host "Pivot CLI Installer for Windows" -ForegroundColor $Colors.White
    Write-Host "================================" -ForegroundColor $Colors.White
    Write-Host ""
    
    # Detect architecture
    $architecture = Get-Architecture
    Write-Info "Detected architecture: $architecture"
    
    # Get latest version
    $version = Get-LatestVersion
    Write-Info "Latest version: $version"
    
    # Install
    Install-Pivot -Architecture $architecture -Version $version -InstallDirectory $InstallDir
    
    # Verify
    Test-Installation -InstallDirectory $InstallDir
}

# Run main function
Main

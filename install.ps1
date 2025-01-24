param (
    [string]$Version = ""
)

# Show usage instructions
function Show-Usage {
    Write-Host "Usage: .\install.ps1 [-Version VERSION]"
    Write-Host "  -Version VERSION  Install specific version (e.g. 0.4.2)"
    Write-Host
    Write-Host "Examples:"
    Write-Host "  .\install.ps1              # Install latest version"
    Write-Host "  .\install.ps1 -Version 0.4.2  # Install version 0.4.2"
    exit 1
}

# Validate parameters
if ($PSBoundParameters.Count -gt 0 -and -not $PSBoundParameters.ContainsKey('Version')) {
    Write-Host "Error: Invalid parameters provided" -ForegroundColor Red
    Show-Usage
}

# Detect architecture
function Get-Architecture {
    if ([Environment]::Is64BitOperatingSystem) {
        return "amd64"
    }
    return "Unknown"
}

# Create temporary directory
$TempDir = Join-Path $env:TEMP ([System.Guid]::NewGuid())
New-Item -ItemType Directory -Path $TempDir | Out-Null

try {
    # Get version information
    if ($Version) {
        # Validate specified version exists
        $Version = $Version -replace '^v', ''
        $ReleaseUrl = "https://api.github.com/repos/belingud/gptcomet/releases/tags/v$Version"
        try {
            $Release = Invoke-RestMethod -Uri $ReleaseUrl
            $Tag = "v$Version"
        }
        catch {
            Write-Host "Error: Version v$Version not found" -ForegroundColor Red
            exit 1
        }
    }
    else {
        # Get latest version if no version specified
        Write-Host "Fetching latest release information..."
        $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/belingud/gptcomet/releases/latest"
        $Version = $Release.tag_name -replace '^v', ''
        $Tag = $Release.tag_name
    }

    # Build download URL
    $Arch = Get-Architecture
    if ($Arch -eq "Unknown") {
        Write-Host "Error: Unsupported architecture" -ForegroundColor Red
        exit 1
    }

    $DownloadUrl = "https://github.com/belingud/gptcomet/releases/download/$Tag/gptcomet_${Version}_windows_$Arch.zip"
    $ZipPath = Join-Path $TempDir "gptcomet.zip"

    # Download file
    Write-Host "Downloading gptcomet version $Version..."
    Write-Host "Download URL: $DownloadUrl"
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath

    # Extract files
    Write-Host "Extracting files..."
    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

    # Create install directory
    $InstallDir = Join-Path $env:LOCALAPPDATA "Programs\gptcomet"
    if (-not (Test-Path $InstallDir)) {
        Write-Host "Creating installation directory: $InstallDir"
        New-Item -ItemType Directory -Path $InstallDir | Out-Null
    }

    # Copy executable
    Write-Host "Installing gptcomet to $InstallDir..."
    Copy-Item -Path (Join-Path $TempDir "gptcomet.exe") -Destination $InstallDir -Force

    # Add to PATH if not already there
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable(
            "Path",
            "$UserPath;$InstallDir",
            "User"
        )
        Write-Host "Added $InstallDir to user PATH"
    }

    Write-Host "Installation completed! gptcomet has been installed to $InstallDir"
    Write-Host "You can now run the following commands:"
    Write-Host "  + gptcomet"
    Write-Host "Please restart your terminal for the PATH changes to take effect"
}
catch {
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}
finally {
    # Clean up
    if (Test-Path $TempDir) {
        Write-Host "Cleaning up temporary files..."
        Remove-Item -Path $TempDir -Recurse -Force
    }
}

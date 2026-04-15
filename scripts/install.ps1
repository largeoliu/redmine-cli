param(
    [string]$InstallDir = "",
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"

$REPO = "largeoliu/redmine-cli"
$ASSET_NAME_PREFIX = "redmine-cli"
$BINARY_NAME = "redmine"

if (-not $InstallDir) {
    $InstallDir = Join-Path $env:USERPROFILE ".local\bin"
}

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
}

function Write-Error-Exit {
    param([string]$Message)
    Write-Host "[ERROR] " -ForegroundColor Red -NoNewline
    Write-Host $Message
    exit 1
}

function Get-OS {
    if ($IsWindows -or ($env:OS -match "Windows")) {
        return "windows"
    } elseif ($IsMacOS) {
        return "darwin"
    } elseif ($IsLinux) {
        return "linux"
    }
    Write-Error-Exit "Unsupported OS"
}

function Get-Arch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch) {
        "X64" { return "amd64" }
        "Arm64" { return "arm64" }
        default { Write-Error-Exit "Unsupported architecture: $arch" }
    }
}

function Get-LatestVersion {
    $latestUrl = "https://github.com/$REPO/releases/latest"
    try {
        $response = Invoke-WebRequest -Uri $latestUrl -Method Head -MaximumRedirection 0 -ErrorAction SilentlyContinue
    } catch {
        $response = $_.Exception.Response
    }
    
    if ($response.Headers.Location) {
        $location = $response.Headers.Location.ToString()
        $version = $location.Split("/")[-1]
        return $version
    }
    
    Write-Error-Exit "Failed to get latest version"
}

function Confirm-Checksum {
    param(
        [string]$Version,
        [string]$ArchiveName,
        [string]$ArchivePath
    )
    
    $checksumsUrl = "https://github.com/$REPO/releases/download/$Version/checksums.txt"
    
    Write-Info "Downloading checksums..."
    try {
        $checksumsContent = Invoke-WebRequest -Uri $checksumsUrl -UseBasicParsing | Select-Object -ExpandProperty Content
    } catch {
        Write-Warn "Could not download checksums, skipping verification"
        return
    }
    
    $expectedHash = ($checksumsContent -split "`n" | Where-Object { $_ -match "\s+$([regex]::Escape($ArchiveName))`$" } | Select-Object -First 1) -replace '\s+.*', ''
    if (-not $expectedHash) {
        Write-Warn "Archive not found in checksums file, skipping verification"
        return
    }
    
    Write-Info "Verifying checksum..."
    $actualHash = (Get-FileHash -Path $ArchivePath -Algorithm SHA256).Hash.ToLower()
    if ($actualHash -ne $expectedHash.ToLower()) {
        Write-Error-Exit "Checksum mismatch! Expected: $expectedHash, Actual: $actualHash"
    }
    
    Write-Info "Checksum verified"
}

function Download-Binary {
    param(
        [string]$Version,
        [string]$OS,
        [string]$Arch
    )
    
    $archiveName = "${ASSET_NAME_PREFIX}_$($Version.Substring(1))_${OS}_${Arch}.zip"
    $downloadUrl = "https://github.com/$REPO/releases/download/$Version/$archiveName"
    
    Write-Info "Downloading $archiveName..."
    
    $tmpDir = New-TemporaryDirectory
    $archivePath = Join-Path $tmpDir $archiveName
    
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -UseBasicParsing
    } catch {
        Write-Error-Exit "Failed to download $archiveName : $_"
    }
    
    Confirm-Checksum -Version $Version -ArchiveName $archiveName -ArchivePath $archivePath
    
    Write-Info "Extracting..."
    
    try {
        Expand-Archive -Path $archivePath -DestinationPath $tmpDir -Force
    } catch {
        Write-Error-Exit "Failed to extract archive: $_"
    }
    
    return $tmpDir
}

function New-TemporaryDirectory {
    $tmpPath = [System.IO.Path]::GetTempPath()
    $tmpDir = [System.IO.Path]::Combine($tmpPath, [System.IO.Path]::GetRandomFileName())
    New-Item -ItemType Directory -Path $tmpDir | Out-Null
    return $tmpDir
}

function Install-Binary {
    param([string]$TmpDir)
    
    if (-not (Test-Path $InstallDir)) {
        Write-Info "Creating install directory: $InstallDir"
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    
    $binaryPath = Join-Path $TmpDir "$BINARY_NAME.exe"
    
    if (-not (Test-Path $binaryPath)) {
        $binaryPath = Join-Path $TmpDir $BINARY_NAME
    }
    
    if (-not (Test-Path $binaryPath)) {
        Write-Error-Exit "Binary not found in archive"
    }
    
    $destPath = Join-Path $InstallDir "$BINARY_NAME.exe"
    Move-Item -Path $binaryPath -Destination $destPath -Force
    
    Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue
}

function Test-PathInEnv {
    $pathDirs = $env:PATH -split ";"
    return $pathDirs -contains $InstallDir
}

function Add-ToPath {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$InstallDir*") {
        $newPath = "$currentPath;$InstallDir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Info "Added $InstallDir to user PATH"
    }
}

$os = Get-OS
$arch = Get-Arch

if (-not $Version) {
    $Version = Get-LatestVersion
}

Write-Info "Installing $BINARY_NAME..."
Write-Info "OS: $os, Arch: $arch, Version: $Version"

$tmpDir = Download-Binary -Version $Version -OS $os -Arch $arch
Install-Binary -TmpDir $tmpDir

Write-Info "Successfully installed $BINARY_NAME to $InstallDir"

if (-not (Test-PathInEnv)) {
    Write-Host ""
    Write-Warn "$InstallDir is not in your PATH"
    Write-Host ""
    
    $addToPath = Read-Host "Would you like to add it to your PATH? (Y/n)"
    if ($addToPath -ne "n" -and $addToPath -ne "N") {
        Add-ToPath
        Write-Host ""
        Write-Info "Please restart your terminal or run: `$env:Path = [System.Environment]::GetEnvironmentVariable('Path','User')"
    }
}

Write-Host ""
Write-Info "Run '$BINARY_NAME --help' to get started"

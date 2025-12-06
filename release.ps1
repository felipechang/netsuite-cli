# Script to create and push a release tag
# Usage: .\release.ps1 [version]
# Example: .\release.ps1 1.0.0

param(
    [string]$Version
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Get version from parameter or prompt
if ([string]::IsNullOrWhiteSpace($Version)) {
    $Version = Read-Host "Enter version number (e.g., 1.0.0)"
}

# Validate version format (basic check)
if ($Version -notmatch '^\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$') {
    Write-Host "Error: Invalid version format. Expected format: X.Y.Z or X.Y.Z-prerelease" -ForegroundColor Red
    Write-Host "Example: 1.0.0 or 1.0.0-beta" -ForegroundColor Yellow
    exit 1
}

# Create tag name
$Tag = "v$Version"

# Check if tag already exists locally
try {
    $null = git rev-parse "$Tag" 2>&1
    Write-Host "Error: Tag $Tag already exists locally" -ForegroundColor Red
    exit 1
} catch {
    # Tag doesn't exist locally, which is what we want
}

# Check if tag exists on remote
$remoteTags = git ls-remote --tags origin "$Tag" 2>&1
if ($remoteTags -match $Tag) {
    Write-Host "Error: Tag $Tag already exists on remote" -ForegroundColor Red
    exit 1
}

# Confirm before proceeding
Write-Host "This will create and push tag: $Tag" -ForegroundColor Yellow
$confirm = Read-Host "Continue? (y/N)"
if ($confirm -notmatch '^[Yy]$') {
    Write-Host "Aborted." -ForegroundColor Yellow
    exit 0
}

# Create the tag
Write-Host "Creating tag $Tag..." -ForegroundColor Cyan
git tag $Tag
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Failed to create tag" -ForegroundColor Red
    exit 1
}

# Push the tag to origin
Write-Host "Pushing tag $Tag to origin..." -ForegroundColor Cyan
git push origin $Tag
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Failed to push tag" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Successfully created and pushed tag $Tag" -ForegroundColor Green
Write-Host "GitHub Actions will now build binaries for all platforms and create a release." -ForegroundColor Green

# Get repository URL for monitoring
try {
    $remoteUrl = git config --get remote.origin.url
    if ($remoteUrl -match 'github\.com[:/]([^/]+/[^/]+)\.git') {
        $repoPath = $matches[1]
        $url = 'https://github.com/' + $repoPath + '/actions'
        Write-Host ('You can monitor the progress at: ' + $url) -ForegroundColor Cyan
    }
} catch {
    # If we can't get the URL, just skip showing the link
}

#!/bin/bash

# Script to create and push a release tag
# Usage: ./release.sh [version]
# Example: ./release.sh 1.0.0

set -e

# Get version from argument or prompt
if [ -z "$1" ]; then
    read -p "Enter version number (e.g., 1.0.0): " VERSION
else
    VERSION="$1"
fi

# Validate version format (basic check)
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
    echo "Error: Invalid version format. Expected format: X.Y.Z or X.Y.Z-prerelease"
    echo "Example: 1.0.0 or 1.0.0-beta"
    exit 1
fi

# Create tag name
TAG="v${VERSION}"

# Check if tag already exists locally
if git rev-parse "$TAG" >/dev/null 2>&1; then
    echo "Error: Tag $TAG already exists locally"
    exit 1
fi

# Check if tag exists on remote
if git ls-remote --tags origin "$TAG" | grep -q "$TAG"; then
    echo "Error: Tag $TAG already exists on remote"
    exit 1
fi

# Confirm before proceeding
echo "This will create and push tag: $TAG"
read -p "Continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

# Create the tag
echo "Creating tag $TAG..."
git tag "$TAG"

# Push the tag to origin
echo "Pushing tag $TAG to origin..."
git push origin "$TAG"

echo ""
echo "✓ Successfully created and pushed tag $TAG"
echo "GitHub Actions will now build binaries for all platforms and create a release."
echo "You can monitor the progress at: https://github.com/$(git config --get remote.origin.url | sed -E 's/.*github.com[:/]([^/]+\/[^/]+)\.git/\1/')/actions"

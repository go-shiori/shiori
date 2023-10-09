#!/bin/bash

# This script is used to check if the swagger files are up to date.

# Check if swag version is correct
CURRENT_SWAG_VERSION=$(swag --version | cut -d " " -f 3)
if [ "$CURRENT_SWAG_VERSION" != "$REQUIRED_SWAG_VERSION" ]; then
    echo "swag version is incorrect. Required version: $REQUIRED_SWAG_VERSION, current version: $CURRENT_SWAG_VERSION"
    exit 1
fi

# Check if the git tree for CWD is clean
if [ -n "$(git status docs/swagger --porcelain)" ]; then
    echo "❌ git tree is not clean. Please commit all changes before running this script."
    git diff
    exit 1
fi

# Check swag comments
make swag-fmt
if [ -n "$(git status docs/swagger --porcelain)" ]; then
    echo "❌ swag comments are not formatted. Please run 'make swag-fmt' and commit the changes."
    git diff
    git checkout -- docs/swagger
    exit 1
fi

# Check swagger documentation
make swagger
if [ -n "$(git status docs/swagger --porcelain)" ]; then
    echo "❌ swagger documentation not updated, please run 'make swagger' and commit the changes."
    git diff
    git checkout -- docs/swagger
    exit 1
fi

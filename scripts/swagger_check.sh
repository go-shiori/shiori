#!/bin/bash

# This script is used to check if the swagger files are up to date.

# Check if swag version is correct
# Read the version from the binary build info, not from `swag --version`:
# swag v1.16.5 and v1.16.6 still report "v1.16.4" there, so that output
# cannot tell which release generated the docs.
CURRENT_SWAG_VERSION=$(go version -m "$(command -v swag)" | awk '$1 == "mod" && $2 == "github.com/swaggo/swag" {print $3}')
if [ "$CURRENT_SWAG_VERSION" != "$REQUIRED_SWAG_VERSION" ]; then
    echo "swag version is incorrect. Required version: $REQUIRED_SWAG_VERSION, current version: $CURRENT_SWAG_VERSION"
    exit 1
fi

# Check if the git tree for CWD is clean
if [ -n "$(git status docs/swagger --porcelain)" ]; then
    echo "❌ git tree is not clean. Please commit all changes before running this script."
    exit 1
fi

# Check swag comments
make swag-fmt
if [ -n "$(git status docs/swagger --porcelain)" ]; then
    echo "❌ swag comments are not formatted. Please run 'make swag-fmt' and commit the changes."
    git checkout -- docs/swagger
    exit 1
fi

# Check swagger documentation
make swagger
if [ -n "$(git status docs/swagger --porcelain)" ]; then
    echo "❌ swagger documentation not updated, please run 'make swagger' and commit the changes."
    git checkout -- docs/swagger
    exit 1
fi

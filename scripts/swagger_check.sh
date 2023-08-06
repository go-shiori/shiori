#!/bin/bash

# This script is used to check if the swagger files are up to date.

# Check if swag version is correct
CURRENT_SWAG_VERSION=$(swag --version | cut -d " " -f 3)
if [ "$CURRENT_SWAG_VERSION" != "$REQUIRED_SWAG_VERSION" ]; then
    echo "swag version is incorrect. Required version: $REQUIRED_SWAG_VERSION, current version: $CURRENT_SWAG_VERSION"
    exit 1
fi

# Check if the git tree for CWD is clean
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ git tree is not clean. Please commit all changes before running this script."
    exit 1
fi

# Check swag comments
make swag-fmt 2> /dev/null
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ swag comments are not formatted. Please run 'make swag-fmt' and commit the changes."
    git reset --hard
    exit 1
fi

# Check swagger documentation
TMPDIR=$(mktemp -d)
SWAGGER_DOCS_PATH=$TMPDIR/swagger make swagger 2> /dev/null

diff -r $SWAGGER_DOCS_PATH $TMPDIR/swagger

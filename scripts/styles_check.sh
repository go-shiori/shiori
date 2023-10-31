#!/bin/bash

# This script is used to check if the style.css file is up to date.

# Check if the git tree for CWD is clean
if [ -n "$(git status internal/view/assets/css --porcelain)" ]; then
    echo "❌ git tree is not clean. Please commit all changes before running this script."
    exit 1
fi

# Check if prettier is ok
if ! bun x prettier internal/view/ --check; then
    echo "❌ code style issues found. Please run 'make styles' and commit the changes."
    exit 1
fi

# Check style.css file
CLEANCSS_OPTS=${CLEANCSS_OPTS} make styles
if [ -n "$(git status internal/view/assets/css --porcelain)" ]; then
    echo "❌ style.css wasn't built from less changes. Please run 'make styles' and commit the changes."
    git checkout -- internal/view/assets/css/
    exit 1
fi

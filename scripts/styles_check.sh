#!/bin/bash

# This script is used to check if the style.css file is up to date.

# Check if the git tree for CWD is clean
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ git tree is not clean. Please commit all changes before running this script."
    exit 1
fi

# Check style.css file
CLEANCSS_OPTS=${CLEANCSS_OPTS} make styles
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ style.css wasn't built from less changes. Please run 'make styles' and commit the changes."
    git reset --hard
    exit 1
fi

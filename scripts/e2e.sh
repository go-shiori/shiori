#!/bin/bash

set -e

TIMEOUT=30m

# Run the e2e tests
echo "Running e2e tests..."

# Disable Ryuk for CI environments to avoid errors due to each test
# using its own container, which can cause race conditions with Ryuk
# due to us being the ones creating/removing the containers.
export TESTCONTAINERS_RYUK_DISABLED=true

export CONTEXT_PATH=$(pwd)

# if gotestfmt is installed, run with it
if [ -x "$(command -v gotestfmt)" ]; then
    go test ./e2e/... -count=1 -v -timeout=${TIMEOUT} -json | gotestfmt ${GOTESTFMT_FLAGS}
else
    go test ./e2e/... -count=1 -v -timeout=${TIMEOUT}
fi

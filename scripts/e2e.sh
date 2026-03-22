#!/bin/bash

set -e

TIMEOUT=30m

# Clean up testcontainers images on exit (regardless of test result)
cleanup() {
    echo "Cleaning up testcontainers images..."
    docker image prune -f --filter "label=org.testcontainers=true" 2>/dev/null || true
    # Remove images built by testcontainers (UUID-named images)
    docker images --format '{{.Repository}}:{{.Tag}}' | grep -E '^[0-9a-f]{8}-' | xargs -r docker rmi 2>/dev/null || true
}
trap cleanup EXIT

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

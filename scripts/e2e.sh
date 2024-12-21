#!/bin/bash

set -e

TIMEOUT=30m

# Run the e2e tests
echo "Running e2e tests..."

export CONTEXT_PATH=$(pwd)

go test ./e2e/... -count=1 -v -timeout=${TIMEOUT}

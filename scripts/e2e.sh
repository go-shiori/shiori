#!/bin/bash

set -e

TIMEOUT=30m

# Run the e2e tests
echo "Running e2e tests..."
cd e2e

go test ./... -count=1 -v -timeout=${TIMEOUT}

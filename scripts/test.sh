#!/bin/bash

# Check if gotestfmt is installed
GOTESTFMT=$(which gotestfmt)
if [[ -z "${GOTESTFMT}" ]]; then
    echo "gotestfmt not found. Using test standard output."
fi

# if gotestfmt is installed, run with it
if [[ -n "${GOTESTFMT}" ]]; then
    set -o pipefail
    go test ./... ${GO_TEST_FLAGS} -json | gotestfmt ${GOTESTFMT_FLAGS}
else
    go test ./... ${GO_TEST_FLAGS}
fi

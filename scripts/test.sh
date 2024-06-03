#!/bin/bash

# Check if gotestfmt is installed
if ! [ -x "$(command -v gotestfmt)" ]; then
    echo "gotestfmt not found. Using test standard output."
fi

# if gotestfmt is installed, run with it
if [ -x "$(command -v gotestfmt)" ]; then
    set -o pipefail
    go test ${SOURCE_FILES} ${GO_TEST_FLAGS} -json | gotestfmt ${GOTESTFMT_FLAGS}
else
    go test ${SOURCE_FILES} ${GO_TEST_FLAGS}
fi

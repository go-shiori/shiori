#!/bin/bash

INPUT=internal/view/assets/less/style.less
OUTPUT=internal/view/assets/css/style.css

# Use bun is installled
if [ -x "$(command -v bun)" ]; then
    bun install
    bun x lessc $INPUT $OUTPUT
    bun x clean-css-cli -o $OUTPUT $OUTPUT
    exit 0
fi

# Default to lessc and cleancss
lessc $INPUT $OUTPUT
cleancss -o $OUTPUT $OUTPUT

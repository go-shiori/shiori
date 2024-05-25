#!/bin/bash

INPUT_STYLECSS=internal/view/assets/less/style.less
OUTPUT_STYLECSS=internal/view/assets/css/style.css

INPUT_ARCHIVECSS=internal/view/assets/less/archive.less
OUTPUT_ARCHIVECSS=internal/view/assets/css/archive.css

# Use bun is installled
if [ -x "$(command -v bun)" ]; then
    bun install
    bun x prettier internal/view/ --write
    bun x lessc $INPUT_STYLECSS $OUTPUT_STYLECSS
    bun x lessc $INPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS
    bun x clean-css-cli $CLEANCSS_OPTS -o $OUTPUT_STYLECSS $OUTPUT_STYLECSS
    bun x clean-css-cli $CLEANCSS_OPTS -o $OUTPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS
    exit 0
fi

# Default to lessc and cleancss
lessc $INPUT_STYLECSS $OUTPUT_STYLECSS
lessc $INPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS
cleancss $CLEANCSS_OPTS -o $OUTPUT_STYLECSS $OUTPUT_STYLECSS
cleancss $CLEANCSS_OPTS -o $OUTPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS

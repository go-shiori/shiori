#!/bin/bash

INPUT_STYLECSS=internal/view/assets/less/style.less
OUTPUT_STYLECSS=internal/view/assets/css/style.css

INPUT_ARCHIVECSS=internal/view/assets/less/archive.less
OUTPUT_ARCHIVECSS=internal/view/assets/css/archive.css

# Use bun is installled
if [ -x "$(command -v bun)" ]; then
    sde -chip-check-disable -- bun install
    sde -chip-check-disable -- bun x prettier internal/view/ --write
    sde -chip-check-disable -- bun x lessc $INPUT_STYLECSS $OUTPUT_STYLECSS
    sde -chip-check-disable -- bun x lessc $INPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS
    sde -chip-check-disable -- bun x clean-css-cli $CLEANCSS_OPTS -o $OUTPUT_STYLECSS $OUTPUT_STYLECSS
    sde -chip-check-disable -- bun x clean-css-cli $CLEANCSS_OPTS -o $OUTPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS
    exit 0
fi

# Default to lessc and cleancss
lessc $INPUT_STYLECSS $OUTPUT_STYLECSS
lessc $INPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS
cleancss $CLEANCSS_OPTS -o $OUTPUT_STYLECSS $OUTPUT_STYLECSS
cleancss $CLEANCSS_OPTS -o $OUTPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS

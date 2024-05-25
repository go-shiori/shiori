#!/bin/bash

INPUT_STYLECSS=internal/view/assets/less/style.less
OUTPUT_STYLECSS=internal/view/assets/css/style.css

INPUT_ARCHIVECSS=internal/view/assets/less/archive.less
OUTPUT_ARCHIVECSS=internal/view/assets/css/archive.css

# Detect support of avx2

if [ -x "$(command -v bun)" ]; then
    if grep -q avx2 /proc/cpuinfo; then
        BUNEXCUTE="bun"
    else
        BUNEXCUTE="sde -chip-check-disable -- bun"
        echo "Your cpu not support avx2 so We use sde for more information please lookat https://github.com/oven-sh/bun/issues/762#issuecomment-1186505847"
    fi
fi

# Use bun is installled
if [ -x "$(command -v bun)" ]; then
    $BUNEXCUTE install
    $BUNEXCUTE x prettier internal/view/ --write
    $BUNEXCUTE x lessc $INPUT_STYLECSS $OUTPUT_STYLECSS
    $BUNEXCUTE x lessc $INPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS
    $BUNEXCUTE x clean-css-cli $CLEANCSS_OPTS -o $OUTPUT_STYLECSS $OUTPUT_STYLECSS
    $BUNEXCUTE x clean-css-cli $CLEANCSS_OPTS -o $OUTPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS
    exit 0
fi

# Default to lessc and cleancss
lessc $INPUT_STYLECSS $OUTPUT_STYLECSS
lessc $INPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS
cleancss $CLEANCSS_OPTS -o $OUTPUT_STYLECSS $OUTPUT_STYLECSS
cleancss $CLEANCSS_OPTS -o $OUTPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS

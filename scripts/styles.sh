#!/bin/bash

INPUT_STYLECSS=internal/view/assets/less/style.less
OUTPUT_STYLECSS=internal/view/assets/css/style.css

INPUT_ARCHIVECSS=internal/view/assets/less/archive.less
OUTPUT_ARCHIVECSS=internal/view/assets/css/archive.css

# Detect support of avx2
BUN="bun"
case `uname -o` in
    GNU/Linux)
    # Detect support of avx2 in linux hosts
    if ! grep -q avx2 /proc/cpuinfo; then
        echo "It seems that your CPU does not support AVX2, if you experience long build times (>1m) ensure that you use bun's baseline builds. More information at https://github.com/oven-sh/bun/issues/67"
    fi
    ;;
esac

# Use bun is installled
if [ -x "$(command -v bun)" ]; then
    $BUN install
    $BUN x prettier internal/view/ --write
    $BUN x lessc $INPUT_STYLECSS $OUTPUT_STYLECSS
    $BUN x lessc $INPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS
    $BUN x clean-css-cli $CLEANCSS_OPTS -o $OUTPUT_STYLECSS $OUTPUT_STYLECSS
    $BUN x clean-css-cli $CLEANCSS_OPTS -o $OUTPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS
    exit 0
fi

# Default to lessc and cleancss
lessc $INPUT_STYLECSS $OUTPUT_STYLECSS
lessc $INPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS
cleancss $CLEANCSS_OPTS -o $OUTPUT_STYLECSS $OUTPUT_STYLECSS
cleancss $CLEANCSS_OPTS -o $OUTPUT_ARCHIVECSS $OUTPUT_ARCHIVECSS

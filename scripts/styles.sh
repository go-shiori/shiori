#!/bin/bash

CMD="lessc"

# Check if bun is installled
if [ -x "$(command -v bun)" ]; then
    bun install
    CMD="bun x lessc -x"
fi

$CMD internal/view/assets/less/style.less internal/view/assets/css/style.css

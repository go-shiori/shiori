#!/bin/bash

# This script is used to check if the swagger files are up to date.

TMPDIR=$(mktemp -d)
SWAGGER_DOCS_PATH=$TMPDIR/swagger make swagger

diff -r $SWAGGER_DOCS_PATH $TMPDIR/swagger

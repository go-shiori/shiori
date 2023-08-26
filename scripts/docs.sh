#!/bin/bash

VENV_PATH="docs/.venv"


# Check if the docs virtualenv is created
if [[ ! -d "${VENV_PATH}" ]]; then
    python3 -m venv ${VENV_PATH}
fi

# Activate the virtualenv
source ${VENV_PATH}/bin/activate

# Install the requirements
pip install -r docs/requirements.txt

# Build the docs
mkdocs build --clean ${MKDOCS_EXTRA_FLAGS} -d ${DOCS_BUILD_PATH}

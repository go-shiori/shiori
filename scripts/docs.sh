#!/bin/bash

VENV_PATH="docs/.venv"
DOCS_ACTION=${DOCS_ACTION:-"build"}

# Check if the docs virtualenv is created
if [[ ! -d "${VENV_PATH}" ]]; then
    python3 -m venv ${VENV_PATH}
fi

# Activate the virtualenv
source ${VENV_PATH}/bin/activate

# Install the requirements
pip install -r docs/requirements.txt

# Execute action based on the command
if [[ "${DOCS_ACTION}" == "build" ]]; then
    mkdocs build --clean ${MKDOCS_EXTRA_FLAGS} -d ${DOCS_BUILD_PATH}
    exit 0
elif [[ "${DOCS_ACTION}" == "publish" ]]; then
    # Build the docs
    mkdocs gh-deploy -d ${DOCS_BUILD_PATH} --clean ${MKDOCS_EXTRA_FLAGS}
    exit 0
fi

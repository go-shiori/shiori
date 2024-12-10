#!/usr/bin/env bash
set -ex

# Check if the shiori_builder builder exists
if [ "$CONTAINER_RUNTIME" == "docker" ]; then
    if [ -z "$($CONTAINER_RUNTIME buildx ls | grep shiori_builder)" ]; then
        echo "Creating shiori_builder builder"
        $CONTAINER_RUNTIME buildx create --use --name shiori_builder
    fi
fi

if [[ -d "dist/shiori_linux_arm_7" ]]; then
    cp -r dist/shiori_linux_arm_7 dist/shiori_linux_armv7
fi

if [[ -d "dist/shiori_linux_amd64_v1" ]]; then
    cp -r dist/shiori_linux_amd64_v1 dist/shiori_linux_amd64
fi

if [[ -d "dist/shiori_linux_arm64_v8.0" ]]; then
    cp -r dist/shiori_linux_arm64_v8.0 dist/shiori_linux_arm64
fi

$CONTAINER_RUNTIME buildx build \
    -f ${CONTAINERFILE_NAME} \
    --platform=${BUILDX_PLATFORMS} \
    --build-arg "ALPINE_VERSION=${CONTAINER_ALPINE_VERSION}" \
    --build-arg "GOLANG_VERSION=${GOLANG_VERSION}" \
    ${CONTAINER_BUILDX_OPTIONS} \
    .

if [ "$CONTAINER_RUNTIME" == "docker" ]; then
    $CONTAINER_RUNTIME buildx rm shiori_builder
fi

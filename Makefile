GO ?= $(shell command -v go 2> /dev/null)
BASH ?= $(shell command -v bash 2> /dev/null)
GOLANG_VERSION := $(shell head -n 4 go.mod | tail -n 1 | cut -d " " -f 2)

# Development
SHIORI_DIR ?= dev-data
SOURCE_FILES ?=./internal/...

# Build
CGO_ENABLED ?= 0
BUILD_TIME := $(shell date -u +%Y%m%d.%H%M%S)
BUILD_HASH := $(shell git describe --tags)
BUILD_TAGS ?= osusergo,netgo,fts5
LDFLAGS += -s -w -X main.version=$(BUILD_HASH) -X main.date=$(BUILD_TIME)

# Build (container)
CONTAINER_RUNTIME := docker
CONTAINERFILE_NAME := Dockerfile
CONTAINER_ALPINE_VERSION := 3.19
BUILDX_PLATFORMS := linux/amd64,linux/arm64,linux/arm/v7

# This is used for local development only, forcing linux to create linux only images but with the arch
# of the running machine. Far from perfect but works.
LOCAL_BUILD_PLATFORM = linux/$(shell go env GOARCH)

# Testing
GO_TEST_FLAGS ?= -v -race -count=1 -tags $(BUILD_TAGS) -covermode=atomic -coverprofile=coverage.out
GOTESTFMT_FLAGS ?=

# Development
GIN_MODE ?= debug
SHIORI_DEVELOPMENT ?= true

# Swagger
SWAG_VERSION := $(shell grep "swaggo/swag" go.mod | cut -d " " -f 2)
SWAGGER_DOCS_PATH ?= ./docs/swagger

# Frontend
CLEANCSS_OPTS ?= --with-rebase

# Common exports
export GOLANG_VERSION
export CONTAINER_RUNTIME
export CONTAINERFILE_NAME
export CONTAINER_ALPINE_VERSION
export BUILDX_PLATFORMS

export SOURCE_FILES

# Help documentatin à la https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@cat Makefile | grep -v '\.PHONY' |  grep -v '\help:' | grep -B1 -E '^[a-zA-Z0-9_.-]+:.*' | sed -e "s/:.*//" | sed -e "s/^## //" |  grep -v '\-\-' | sed '1!G;h;$$!d' | awk 'NR%2{printf "\033[36m%-30s\033[0m",$$0;next;}1' | sort

## Cleans up build artifacts
.PHONY: clean
clean:
	rm -rf dist

## Runs the legacy http API for local development
.PHONY: serve
serve:
	SHIORI_DEVELOPMENT=$(SHIORI_DEVELOPMENT) SHIORI_DIR=$(SHIORI_DIR) go run main.go serve

## Runs server for local development
.PHONY: run-server
run-server:
	GIN_MODE=$(GIN_MODE) SHIORI_DEVELOPMENT=$(SHIORI_DEVELOPMENT) SHIORI_DIR=$(SHIORI_DIR) SHIORI_HTTP_SECRET_KEY=shiori go run main.go server --log-level debug

## Generate swagger docs
.PHONY: swagger
swagger:
	SWAGGER_DOCS_PATH=$(SWAGGER_DOCS_PATH) $(BASH) ./scripts/swagger.sh

.PHONY: swag-check
swag-check:
	REQUIRED_SWAG_VERSION=$(SWAG_VERSION) SWAGGER_DOCS_PATH=$(SWAGGER_DOCS_PATH) $(BASH) ./scripts/swagger_check.sh

.PHONY: swag-fmt
swag-fmt:
	swag fmt --dir internal/http
	go fmt ./internal/http/...

## Run linters
.PHONY: lint
lint: golangci-lint swag-check

## Run golangci-lint
.PHONY: golangci-lint
golangci-lint:
	golangci-lint run

## Run unit tests
.PHONY: unittest
unittest:
	GIN_MODE=$(GIN_MODE) GO_TEST_FLAGS="$(GO_TEST_FLAGS)" GOTESTFMT_FLAGS="$(GOTESTFMT_FLAGS)" $(BASH) -xe ./scripts/test.sh

## Run end to end tests
.PHONY: e2e
e2e:
	$(BASH) -xe ./scripts/e2e.sh

## Build styles
.PHONY: styles
styles:
	CLEANCSS_OPTS=$(CLEANCSS_OPTS) $(BASH) ./scripts/styles.sh

## Build styles
.PHONY: styles-check
styles-check:
	CLEANCSS_OPTS=$(CLEANCSS_OPTS) $(BASH) ./scripts/styles_check.sh

## Build binary
.PHONY: build
build: clean
	GIN_MODE=$(GIN_MODE) goreleaser build --clean --snapshot

## Build binary for current targer
build-local: clean
	GIN_MODE=$(GIN_MODE) goreleaser build --clean --snapshot --single-target

## Build docker image using Buildx.
# used for multi-arch builds suing mainly the CI, that's why the task does not
# build the binaries using a dependency task.
.PHONY: buildx
buildx:
	$(info: Make: Buildx)
	@bash scripts/buildx.sh

## Build docker image for local development
buildx-local: build-local
	$(info: Make: Build image locally)
	CONTAINER_BUILDX_OPTIONS="-t shiori:localdev --output type=docker" BUILDX_PLATFORMS=$(LOCAL_BUILD_PLATFORM) scripts/buildx.sh

## Creates a coverage report
.PHONY: coverage
coverage:
	$(GO) test $(GO_TEST_FLAGS) -coverprofile=coverage.txt $(SOURCE_FILES)
	$(GO) tool cover -html=coverage.txt

## Run generate accross the project
.PHONY: generated
generate:
	$(GO) generate ./...

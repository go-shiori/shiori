GO ?= $(shell command -v go 2> /dev/null)
BASH ?= $(shell command -v bash 2> /dev/null)

# Development
SHIORI_DIR ?= dev-data

# Testing
GO_TEST_FLAGS ?= -v -race
GOTESTFMT_FLAGS ?=

# Build
CGO_ENABLED ?= 0
BUILD_TIME := $(shell date -u +%Y%m%d.%H%M%S)
BUILD_HASH := $(shell git describe --tags)
BUILD_TAGS ?= osusergo,netgo
LDFLAGS += -s -w -X main.version=$(BUILD_HASH) -X main.date=$(BUILD_TIME)

# Development
GIN_MODE ?= debug
SHIORI_DEVELOPMENT ?= true

# Swagger
SWAGGER_DOCS_PATH ?= ./docs/swagger

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
	GIN_MODE=$(GIN_MODE) SHIORI_DEVELOPMENT=$(SHIORI_DEVELOPMENT) SHIORI_DIR=$(SHIORI_DIR) go run main.go server

## Generate swagger docs
.PHONY: swagger
swagger:
	SWAGGER_DOCS_PATH=$(SWAGGER_DOCS_PATH) $(BASH) -xe ./scripts/swagger.sh

.PHONY: swagger-check
swagger-check:
	SWAGGER_DOCS_PATH=$(SWAGGER_DOCS_PATH) $(BASH) -xe ./scripts/swagger_check.sh

## Run linter
.PHONY: lint
lint:
	golangci-lint run

## Run unit tests
.PHONY: unittest
unittest:
	GIN_MODE=$(GIN_MODE) GO_TEST_FLAGS="$(GO_TEST_FLAGS)" GOTESTFMT_FLAGS="$(GOTESTFMT_FLAGS)" $(BASH) -xe ./scripts/test.sh

## Build binary
.PHONY: build
build: clean
	GIN_MODE=$(GIN_MODE) goreleaser build --rm-dist --snapshot

## Creates a coverage report
.PHONY: coverage
coverage:
	$(GO) test $(GO_TEST_FLAGS) -coverprofile=coverage.txt ./...
	$(GO) tool cover -html=coverage.txt

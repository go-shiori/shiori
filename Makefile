GO ?= $(shell command -v go 2> /dev/null)
BASH ?= $(shell command -v bash 2> /dev/null)

# Development
SHIORI_DIR ?= dev-data

# Testing
GO_TEST_FLAGS ?= -v -race -count=1
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
SWAG_VERSION := $(shell grep "swaggo/swag" go.mod | cut -d " " -f 2)
SWAGGER_DOCS_PATH ?= ./docs/swagger

# Frontend
CLEANCSS_OPTS ?= --with-rebase

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
	GIN_MODE=$(GIN_MODE) SHIORI_DEVELOPMENT=$(SHIORI_DEVELOPMENT) SHIORI_DIR=$(SHIORI_DIR) SHIORI_HTTP_SECRET_KEY=shiori go run main.go server

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
	GIN_MODE=$(GIN_MODE) goreleaser build --rm-dist --snapshot

## Creates a coverage report
.PHONY: coverage
coverage:
	$(GO) test $(GO_TEST_FLAGS) -coverprofile=coverage.txt ./...
	$(GO) tool cover -html=coverage.txt

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
CONTAINER_ALPINE_VERSION := 3.22
BUILDX_PLATFORMS := linux/amd64,linux/arm64,linux/arm/v7

# This is used for local development only, forcing linux to create linux only images but with the arch
# of the running machine. Far from perfect but works.
LOCAL_BUILD_PLATFORM = linux/$(shell go env GOARCH)

# Testing
GO_TEST_FLAGS ?= -v -race -count=1 -tags $(BUILD_TAGS) -covermode=atomic -coverprofile=coverage.out
GOTESTFMT_FLAGS ?=
SHIORI_TEST_MYSQL_URL ?=shiori:shiori@tcp(127.0.0.1:3306)/shiori
SHIORI_TEST_MARIADB_URL ?= shiori:shiori@tcp(127.0.0.1:3307)/shiori
SHIORI_TEST_PG_URL ?= postgres://shiori:shiori@127.0.0.1:5432/shiori?sslmode=disable

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

export SHIORI_TEST_MYSQL_URL
export SHIORI_TEST_MARIADB_URL
export SHIORI_TEST_PG_URL
export SHIORI_DIR

export SHIORI_TEST_PG_URL=postgres://shiori:shiori@127.0.0.1:5432/shiori?sslmode=disable
export SHIORI_TEST_MARIADB_URL=shiori:shiori@tcp(127.0.0.1:3307)/shiori
export SHIORI_TEST_MYSQL_URL=shiori:shiori@tcp(127.0.0.1:3306)/shiori
export SHIORI_HTTP_SECRET_KEY=shiori

# Help documentatin à la https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@cat Makefile | grep -v '\.PHONY' |  grep -v '\help:' | grep -B1 -E '^[a-zA-Z0-9_.-]+:.*' | sed -e "s/:.*//" | sed -e "s/^## //" |  grep -v '\-\-' | sed '1!G;h;$$!d' | awk 'NR%2{printf "\033[36m%-30s\033[0m",$$0;next;}1' | sort

## Cleans up build artifacts
.PHONY: clean
clean:
	rm -rf dist

## Runs server for local development
.PHONY: run-server
run-server: generate
	HIORI_DEVELOPMENT=$(SHIORI_DEVELOPMENT) SHIORI_HTTP_CORS_ENABLED=true SHIORI_HTTP_CORS_ORIGINS=http://127.0.0.1:5173,http://localhost:5173,http://localhost:8080 SHIORI_HTTP_SERVE_SWAGGER=true go run main.go server --log-level debug

## Runs server for local development with v2 web UI
.PHONY: run-server-v2
run-server-v2: generate
	SHIORI_DEVELOPMENT=$(SHIORI_DEVELOPMENT) SHIORI_HTTP_SERVE_SWAGGER=true SHIORI_HTTP_SERVE_WEB_UI_V2=true go run main.go server --log-level debug

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
	GO_TEST_FLAGS="$(GO_TEST_FLAGS)" GOTESTFMT_FLAGS="$(GOTESTFMT_FLAGS)" $(BASH) -xe ./scripts/test.sh

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
	goreleaser build --clean --snapshot

## Build binary for current targer
build-local: clean
	goreleaser build --clean --snapshot --single-target

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

## Generate TypeScript client
.PHONY: generate-client
generate-client: swagger
	rm -rf ./clients/ts
	openapi-generator generate -i $(SWAGGER_DOCS_PATH)/swagger.json -g typescript-fetch -o ./clients/ts --skip-validate-spec \
		--additional-properties=typescriptThreePlus=true,supportsES6=true,npmName=shiori-api,npmVersion=1.0.0,modelPropertyNaming=original

## Build TypeScript client to JavaScript for frontend
.PHONY: build-client
build-client: generate-client
	cd ./clients/ts && bun install && bun run build
	mkdir -p ./internal/view/assets/js/client
	cd ./clients/ts && bun build ./wrapper.js --outfile=../../internal/view/assets/js/client/shiori-api.js --format=iife

## Build Vue webapp
.PHONY: build-webapp
build-webapp: generate-client
	cd webapp && bun install && bun run build

## Run Vue webapp dev server
.PHONY: run-webapp
run-webapp: generate-client
	cd webapp && bun install && VITE_API_BASE_URL=http://localhost:8080 bun run dev

## Run generate accross the project
.PHONY: generate
generate:
	$(GO) generate ./...

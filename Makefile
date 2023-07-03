SHIORI_DIR ?= dev-data

.PHONY: run-server
run-server: ### Run server
	SHIORI_DEVELOPMENT=true SHIORI_DIR=$(SHIORI_DIR) go run main.go server

.PHONY: swag
swag: ### Generate swagger docs
	swag init

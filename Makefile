.PHONY: all
all: build

.PHONY: build
build:
	bin/build

.PHONY: docker_build
docker_build:
	bin/docker/build

.PHONY: setup
setup:
	bin/setup

.PHONY: test
setup:
	bin/test

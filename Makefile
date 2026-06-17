GO_IMAGE ?= golang:1.26.4-alpine
TASK_VERSION ?= latest
WORKDIR := /workspace
GO_BIN_VOLUME ?= docktui-go-bin
GO_MOD_VOLUME ?= docktui-go-mod
GO_BUILD_VOLUME ?= docktui-go-build

export GO_IMAGE
export GO_BIN_VOLUME
export GO_MOD_VOLUME
export GO_BUILD_VOLUME

DOCKER_RUN := docker compose run --rm -T app
DOCKER_RUN_TTY := docker compose run --rm app

TASK := sh -lc 'export PATH=/go/bin:/usr/local/go/bin:$$PATH; go install github.com/go-task/task/v3/cmd/task@$(TASK_VERSION) && task "$$@"' task

export UID := $(shell id -u)
export GID := $(shell id -g)
export PWD := $(shell pwd)
export FMT_IMAGE ?= ghcr.io/oullin/go-fmt:v0.4.1

FMT_RUN := docker compose run --rm -T fmt

.PHONY: run fmt format-all lint test build smoke

run:
	$(DOCKER_RUN_TTY) $(TASK) run

fmt:
	$(FMT_RUN) format .

format-all:
	$(FMT_RUN) format .

lint:
	$(DOCKER_RUN) $(TASK) lint

test:
	$(DOCKER_RUN) $(TASK) test

build:
	$(DOCKER_RUN) $(TASK) build

smoke:
	$(DOCKER_RUN) $(TASK) smoke

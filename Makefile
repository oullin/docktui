GO_IMAGE ?= golang:1.26.4-alpine
TASK_VERSION ?= latest
WORKDIR := /workspace
GO_BIN_VOLUME ?= docktui-go-bin
GO_MOD_VOLUME ?= docktui-go-mod
GO_BUILD_VOLUME ?= docktui-go-build

DOCKER_RUN := docker run --rm \
	-v "$(PWD)":$(WORKDIR) \
	-w $(WORKDIR) \
	-v $(GO_BIN_VOLUME):/go/bin \
	-v $(GO_MOD_VOLUME):/go/pkg/mod \
	-v $(GO_BUILD_VOLUME):/root/.cache/go-build \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-e DOCKER_HOST=unix:///var/run/docker.sock \
	$(GO_IMAGE)

DOCKER_RUN_TTY := docker run --rm -it \
	-v "$(PWD)":$(WORKDIR) \
	-w $(WORKDIR) \
	-v $(GO_BIN_VOLUME):/go/bin \
	-v $(GO_MOD_VOLUME):/go/pkg/mod \
	-v $(GO_BUILD_VOLUME):/root/.cache/go-build \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-e DOCKER_HOST=unix:///var/run/docker.sock \
	$(GO_IMAGE)

TASK := sh -lc 'export PATH=/go/bin:/usr/local/go/bin:$$PATH; go install github.com/go-task/task/v3/cmd/task@$(TASK_VERSION) && task "$$@"' task

.PHONY: run fmt format-all lint test build smoke

run:
	$(DOCKER_RUN_TTY) $(TASK) run

fmt:
	$(DOCKER_RUN) $(TASK) fmt

format-all:
	$(DOCKER_RUN) $(TASK) format-all

lint:
	$(DOCKER_RUN) $(TASK) lint

test:
	$(DOCKER_RUN) $(TASK) test

build:
	$(DOCKER_RUN) $(TASK) build

smoke:
	$(DOCKER_RUN) $(TASK) smoke

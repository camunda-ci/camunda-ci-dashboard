#!/usr/bin/make
.DEFAULT_GOAL:=help

BINARY=camunda-ci-dashboard
IMAGE_NAME=gcr.io/ci-30-162810/camunda-ci-dashboard:latest
DYNAMIC_COMPILE=0

SHELL:=/bin/bash
SOURCE_DIR=.
PROGRAM_DIR=./cmd
ASSETS:=cmd/bindata_assetfs.go

RUN_CMD=./bin/$(BINARY)
RUN_CMD_OPTS=

GO_BINDATA_OPTS=
GO_BUILD_DEBUG_OPTS=

DEBUG?=false
ifeq ($(DEBUG), true)
	GO_BINDATA_OPTS=$(GO_BINDATA_OPTS) -debug
	GO_BUILD_DEBUG_OPTS=$(GO_BUILD_DEBUG_OPTS) -race
	RUN_CMD_OPTS=$(RUN_CMD_OPTS) --debug=true
endif

# User-friendly check for go
ifeq ($(shell which go &>/dev/null; echo $$?), 0)
	GO_PACKAGES=$(shell go list $(SOURCE_DIR)/... | grep -v vendor)
else
	ECHO=$(shell echo "The 'go' command was not found.")
endif

LDFLAGS=-ldflags "-w -s -X github.com/camunda-ci/camunda-ci-dashboard/cmd/main.Build=`git rev-parse HEAD`"
GO_BUILD_OPTS=-installsuffix cgo $(LDFLAGS) -o bin/$(BINARY) $(GO_BUILD_DEBUG_OPTS)
GO_BUILD_CMD=go build $(GO_BUILD_OPTS) $(PROGRAM_DIR)

ENV_FILE=dashboard-env.sh

# check for environment file
ifneq ($(wildcard $(ENV_FILE)),)
	RUN_CMD=. ./$(ENV_FILE) && ./bin/$(BINARY)
endif

.PHONY: prerequisites
prerequisites: ## pull required prerequisites
	go get -u github.com/golang/lint/golint
	go get -u github.com/kardianos/govendor
	go get -u github.com/jteeuwen/go-bindata/...
	go get -u github.com/elazarl/go-bindata-assetfs/...

.PHONY: clean
clean: ## clean everything
	rm -rf bin/
	mkdir bin
	go clean -i -x

.PHONY: generate-assets
generate-assets: ## include website as assets
	go-bindata-assetfs $(GO_BINDATA_OPTS) assets/... -o bindata_assetfs.go
	mv bindata_assetfs.go $(ASSETS)

.PHONY: deps
deps: ## pull dependencies
	govendor sync

.PHONY: lint
lint: ## lint files
	gofmt -s -l -w $$(find $(SOURCE_DIR) -name "*.go" | grep -v vendor | uniq)
	go list ./... | grep -v /vendor/ | xargs -L1 golint

.PHONY: vet
vet: ## run go vet
	go vet $(GO_PACKAGES)

.PHONY: test
test: ## run tests
	go test -v -cover -p=1 $(GO_PACKAGES)

.PHONY: run
run: generate-assets ## run dashboard without compilation (env file required)
	. ./dashboard-env.sh && go run ./cmd/*.go $(RUN_CMD_OPTS)

.PHONY: run-binary
run-binary: build ## directly build and execute binary
	$(RUN_CMD) $(RUN_CMD_OPTS)

.PHONY: local
local: clean generate-assets lint vet test compile ## run everything (Inet not required)

.PHONY: build
build: clean prerequisites deps generate-assets lint vet test compile ## run everything (Inet required)

.PHONY: compile
compile: ## compile for local machine
	CGO_ENABLED=$(DYNAMIC_COMPILE) $(GO_BUILD_CMD) && chmod u+x bin/$(BINARY)

.PHONY: test-compile
test-compile: ## compile for local machine with race condition detection
	go build $(SOURCE_DIR) -v -x -race $(SOURCE_DIR)

.PHONY: distribution
distribution: ## compile linux_amd64 binary
	rm bin/*
	CGO_ENABLED=$(DYNAMIC_COMPILE) GOOS=linux GOARCH=amd64 $(GO_BUILD_CMD)
	chmod u+x bin/$(BINARY) && mv bin/$(BINARY) bin/$(BINARY)_linux_amd64

# DOCKER targets
.PHONY: package
package: ## build docker image (Go compilation will be done in MultiStage Dockerfile)
	docker build --pull -t $(IMAGE_NAME) .

.PHONY: push
push: ## push docker image
	docker push $(IMAGE_NAME)

.PHONY: docker-run
docker-run: ## start camunda-ci-dashboard locally
	docker run -it --rm -p 8000:8000 $(IMAGE_NAME)

.PHONY: help
help: ## show targets with comments
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


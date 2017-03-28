#!/usr/bin/make

SHELL := /bin/bash
GO:=$(shell which go)
SOURCE_DIR=.
PROGRAM_DIR=./cmd
ASSETS:=cmd/bindata_assetfs.go
GO_PACKAGES = $(shell go list ./... | grep -v vendor)
GO_FILES = $(shell find $(SOURCE_DIR) -name "*.go" | grep -v vendor | uniq)

BINARY=camunda-ci-dashboard

LDFLAGS=-ldflags "-X github.com/camunda-ci/camunda-ci-dashboard/cmd/main.Build=`git rev-parse HEAD`"
GO_BUILD_CMD=go build -a -installsuffix cgo ${LDFLAGS} -o bin/${BINARY} -v -x $(PROGRAM_DIR)

IMAGE_NAME=registry.camunda.com/camunda-ci-dashboard:latest

RUN=./bin/$(BINARY)
ENV_FILE=dashboard-env.sh

# check for environment file
ifneq ($(wildcard $(ENV_FILE)),)
	RUN=. ./dashboard-env.sh && ./bin/$(BINARY)
endif

.DEFAULT_GOAL: build

build: prerequisites clean deps generate-assets lint test compile

build-debug: prerequisites clean deps generate-debug-assets lint test compile

compile:
	${GO_BUILD_CMD} && chmod u+x bin/${BINARY}

distribution:
	rm bin/*
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ${GO_BUILD_CMD} && chmod u+x bin/${BINARY} && mv bin/${BINARY} bin/${BINARY}_linux_amd64

docker-build: build distribution
	docker build -t $(IMAGE_NAME) .

docker-stage:
	docker run -it --rm -p 8000:8000 $(IMAGE_NAME)

docker-push:
	docker push $(IMAGE_NAME)

docker-publish: docker-build docker-push

prerequisites:
	go get -u github.com/golang/lint/golint
	go get -u github.com/kardianos/govendor
	go get -u github.com/jteeuwen/go-bindata/...
	go get github.com/elazarl/go-bindata-assetfs/...

clean:
	rm -rf bin/ && mkdir -p bin
	go clean -i -x

generate-assets:
	go-bindata-assetfs assets/...
	mv bindata_assetfs.go $(ASSETS)

generate-debug-assets:
	go-bindata-assetfs -debug assets/...
	mv bindata_assetfs.go $(ASSETS)

lint:
	gofmt -s -l -w $(GO_FILES)
	go list ./... | grep -v /vendor/ | xargs -L1 golint

deps:
	govendor sync

vet:
	go vet $(GO_PACKAGES)

test: vet
	go test -v -cover -p=1 $(GO_PACKAGES)

test-compile:
	go build $(SOURCE_DIR) -v -x -race $(SOURCE_DIR)

run: generate-debug-assets
	#go run ./cmd/*.go -debug
	. ./dashboard-env.sh && go run ./cmd/*.go --debug=true

run-binary: build
	$(RUN)

run-binary-debug: build-debug
	$(RUN) --debug=true

.PHONY: build build-debug clean compile deps distribution docker docker-stage docker-publish docker-push generate-assets generate-debug-assets lint prerequisites run run-binary run-binary-debug test test-compile vet

#!/usr/bin/make

SHELL := /bin/bash
GO:=$(shell which go)
SOURCE_DIR=.
PROGRAM_DIR=./cmd
ASSETS:=cmd/bindata_assetfs.go
GO_PACKAGES = $(shell go list ./... | grep -v vendor)
GO_FILES = $(shell find $(SOURCE_DIR) -name "*.go" | grep -v vendor | uniq)
DYNAMIC_COMPILE=0

BINARY=camunda-ci-dashboard

LDFLAGS=-ldflags "-w -s -X github.com/camunda-ci/camunda-ci-dashboard/cmd/main.Build=`git rev-parse HEAD`"
GO_BUILD_CMD=go build -installsuffix cgo ${LDFLAGS} -o bin/${BINARY} -v -x $(PROGRAM_DIR)

IMAGE_NAME=registry.camunda.com/camunda-ci-dashboard:latest

# https://github.com/upx/upx/releases/download/v3.93/upx-3.93-amd64_linux.tar.xz
#

RUN=./bin/$(BINARY)
ENV_FILE=dashboard-env.sh

# check for environment file
ifneq ($(wildcard $(ENV_FILE)),)
	RUN=. ./dashboard-env.sh && ./bin/$(BINARY)
endif

USE_UPX=false
ifeq ($(shell uname), Linux)
	USE_UPX=true
endif

.DEFAULT_GOAL: build

build-offline: clean generate-assets lint test compile

build: clean prerequisites deps generate-assets lint test compile

build-debug: prerequisites clean deps generate-debug-assets lint test compile

compile:
	CGO_ENABLED=${DYNAMIC_COMPILE} ${GO_BUILD_CMD} && chmod u+x bin/${BINARY}

distribution:
	rm bin/*
	CGO_ENABLED=${DYNAMIC_COMPILE} GOOS=linux GOARCH=amd64 ${GO_BUILD_CMD} && chmod u+x bin/${BINARY} && mv bin/${BINARY} bin/${BINARY}_linux_amd64
	@if [ "${USE_UPX}" == "true" ]; then \
		upx/upx -f --brute bin/${BINARY}_linux_amd64; \
	fi

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
	@if [ "${USE_UPX}" == "true" ]; then \
		curl -sSL https://github.com/upx/upx/releases/download/v3.94/upx-3.94-amd64_linux.tar.xz | tar -xvf - --strip=1 -C upx && chmod +x -R upx; \
	fi

clean:
	rm -rf bin/ && mkdir -p bin
	rm -rf upx/ && mkdir -p upx
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

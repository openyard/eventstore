########################################################## #
# Makefile for Golang Project
# Includes cross-compiling, installation, cleanup
# ########################################################## #

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd docker oapi-codegen go-bindata
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH")))

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

# Determine commands by looking into cmd/*
COMMANDS=$(wildcard ${ROOT_DIR}/cmd/*)
# Determine binary names by stripping out the dir names
BINS=$(foreach cmd,${COMMANDS},$(notdir ${cmd}))
BINARY=$(basename $(ROOT_DIR))
DOCKERFILES=$(foreach dockerfile,$(wildcard ${ROOT_DIR}/build/package/docker/*),$(notdir ${dockerfile}))
TEST_BINARY=eventstore-test
VERSION=TRUNK
BUILD=`git rev-parse HEAD`
BUILD_TAG=dev.`git rev-parse --short HEAD`
PLATFORMS=linux
ARCHITECTURES=amd64
PROJECT_PATH=$(shell pwd)

# Setup linker flags option for build that interoperate with variable names in src code
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION} -X main.build=${BUILD}"
# Use Golang network stack
NETGO=-installsuffix netgo -tags netgo

default: all

all: clean fmt compile test build build-all

# Remove only what we've created
clean:
	if [ -d "$(PROJECT_PATH)/pkg/genproto" ]; then find "$(PROJECT_PATH)/pkg/genproto" -name '*.pb.go' -type f -delete ; fi
	if [ -d "$(PROJECT_PATH)/.coverage" ]; then rm -r "$(PROJECT_PATH)/.coverage" ; fi
	find bin/ -name '*' -type f -delete

compile:
	go generate ./...

tools:
	go get -tool github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	go get -tool github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	go get -tool google.golang.org/protobuf/cmd/protoc-gen-go
	go get -tool google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go install tool

list:
	@grep '^[^#[:space:]].*:' Makefile

fmt:
	go fmt ./...

test:
	go test -count=1 -short -race -v ./...

build:
	mkdir -p bin
	$(foreach BINARY, $(BINS), $(shell go build -v ${LDFLAGS} ${NETGO} -o bin/${BINARY} ./cmd/${BINARY}))

build-all:
	mkdir -p bin
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), \
	$(foreach BINARY, $(BINS), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); go build -v ${LDFLAGS} ${NETGO} -o bin/$(BINARY)-$(GOOS)-$(GOARCH) ./cmd/${BINARY}/))))

package:
	$(foreach DOCKERFILE, $(DOCKERFILES), $(shell docker build -t openyard/eventstore-${VERSION}-${BUILD_TAG} -f $(DOCKERFILE) .))

.PHONY: all clean fmt test compile build build-all

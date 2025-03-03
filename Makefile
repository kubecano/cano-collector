PACKAGE=github.com/kubecano/cano-collector
CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist
CLI_NAME=cano-collector
BIN_NAME=cano-collector
CGO_FLAG=0

HOST_OS:=$(shell go env GOOS)
HOST_ARCH:=$(shell go env GOARCH)

TARGET_ARCH?=linux/amd64

VERSION=$(shell cat ${CURRENT_DIR}/VERSION)

CANO_LINT_GOGC?=20

.PHONY: gogen
gogen:
	export GO111MODULE=off
	go generate ./...

.PHONY: mod-download
mod-download:
	go mod download && go mod tidy # go mod download changes go.sum https://github.com/golang/go/issues/42970

.PHONY: mod-vendor
mod-vendor: mod-download
	go mod vendor

# Run linter on the code (local version)
.PHONY: lint
lint:
	golangci-lint --version
	# NOTE: If you get a "Killed" OOM message, try reducing the value of GOGC
	# See https://github.com/golangci/golangci-lint#memory-usage-of-golangci-lint
	GOGC=$(CANO_LINT_GOGC) GOMAXPROCS=2 golangci-lint run --fix --verbose

.PHONY: vet
vet:
	go vet -json ./...

# Build all Go code (local version)
.PHONY: build
build:
	go build -v `go list ./...`

# Run all unit tests (local version)
.PHONY: test
test:
	go test -v `go list ./...`

.PHONY: help
help:
	@echo 'Common targets'
	@echo
	@echo 'build:'
	@echo '  build(-local)             -- compile go'

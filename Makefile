#! /usr/bin/make -f

# Project variables.
PROJECT_NAME = gex
DATE := $(shell date '+%Y-%m-%dT%H:%M:%S')
HEAD = $(shell git rev-parse HEAD)
LD_FLAGS = 
BUILD_FLAGS = -mod=readonly -ldflags='$(LD_FLAGS)'
BUILD_FOLDER = ./dist

## install: Install de binary.
install:
	@echo Installing Gex...
	@go install $(BUILD_FLAGS) ./...
	@ignite version

## build: Build the binary.
build:
	@echo Building Gex...
	@-mkdir -p $(BUILD_FOLDER) 2> /dev/null
	@go build $(BUILD_FLAGS) -o $(BUILD_FOLDER) ./...

## prepare snapcraft config for release
snapcraft:
	@sed -i 's/{{version}}/'$(version)'/' packaging/snap/snapcraft.yaml

## clean: Clean build files. Also runs `go clean` internally.
clean:
	@echo Cleaning build cache...
	@-rm -rf $(BUILD_FOLDER) 2> /dev/null
	@go clean ./...

.PHONY: install build clean

## govet: Run go vet.
govet:
	@echo Running go vet...
	@go vet ./...

## govulncheck: Run govulncheck
govulncheck:
	@echo Running govulncheck...
	@go run golang.org/x/vuln/cmd/govulncheck ./...

## format: Install and run goimports and gofumpt
format:
	@echo Formatting...
	@go run mvdan.cc/gofumpt -w .
	@go run golang.org/x/tools/cmd/goimports -w -local github.com/ignite/gex .

## lint: Run Golang CI Lint.
lint:
	@echo Running golangci-lint...
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --out-format=tab --issues-exit-code=0

lint-fix:
	@echo Running golangci-lint...
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix --out-format=tab --issues-exit-code=0

.PHONY: govet format lint

## test-unit: Run the unit tests.
test-unit:
	@echo Running unit tests...
	@go test -race -failfast -v ./...

## test: Run unit and tests.
test: govet govulncheck test-unit

.PHONY: test-unit test

help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECT_NAME)", or just run 'make' for install"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

.PHONY: help

.DEFAULT_GOAL := install

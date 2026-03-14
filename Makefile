BINARY   := skills-cli
MODULE   := github.com/darrenrowley/skills-cli
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE     ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  := -ldflags "-X $(MODULE)/cmd.Version=$(VERSION) -X $(MODULE)/cmd.Commit=$(COMMIT) -X $(MODULE)/cmd.BuildDate=$(DATE)"

.PHONY: build test lint clean install help

## build: compile the binary to ./skills-cli
build:
	go build $(LDFLAGS) -o $(BINARY) .

## install: install the binary to $GOPATH/bin
install:
	go install $(LDFLAGS) .

## test: run all tests
test:
	go test -race ./...

## test-cover: run tests with coverage report
test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: run golangci-lint (must be installed)
lint:
	golangci-lint run ./...

## clean: remove build artifacts
clean:
	rm -f $(BINARY) coverage.out coverage.html

## help: show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'

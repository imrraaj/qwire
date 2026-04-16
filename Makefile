GO ?= go
GOCACHE ?= /tmp/go-build
BUILD_DIR ?= build
MAIN_BINARY ?= $(BUILD_DIR)/binprot
SERVER_BINARY ?= $(BUILD_DIR)/server
CLIENT_BINARY ?= $(BUILD_DIR)/client

.PHONY: help fmt test test-race test-cover vet build build-main build-server build-client run run-server run-client clean

help:
	@printf '%s\n' \
		'make fmt         Format Go source files' \
		'make test        Run unit and integration tests' \
		'make test-race   Run tests with the race detector' \
		'make test-cover  Run tests with coverage output' \
		'make vet         Run go vet' \
		'make build       Build all demo binaries into build/' \
		'make run         Run the root demo program' \
		'make run-server  Run the TCP server demo' \
		'make run-client  Run the TCP client demo' \
		'make clean       Remove generated build artifacts'

fmt:
	$(GO) fmt ./...

test:
	GOCACHE=$(GOCACHE) $(GO) test ./...

test-race:
	GOCACHE=$(GOCACHE) $(GO) test -race ./...

test-cover:
	GOCACHE=$(GOCACHE) $(GO) test -cover ./...

vet:
	GOCACHE=$(GOCACHE) $(GO) vet ./...

build: build-main build-server build-client

build-main:
	mkdir -p $(BUILD_DIR)
	GOCACHE=$(GOCACHE) $(GO) build -o $(MAIN_BINARY) .

build-server:
	mkdir -p $(BUILD_DIR)
	GOCACHE=$(GOCACHE) $(GO) build -o $(SERVER_BINARY) ./server

build-client:
	mkdir -p $(BUILD_DIR)
	GOCACHE=$(GOCACHE) $(GO) build -o $(CLIENT_BINARY) ./client

run:
	GOCACHE=$(GOCACHE) $(GO) run .

run-server:
	GOCACHE=$(GOCACHE) $(GO) run ./server

run-client:
	GOCACHE=$(GOCACHE) $(GO) run ./client

clean:
	rm -rf $(BUILD_DIR) coverage.out

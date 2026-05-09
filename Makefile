GO ?= go
APP_NAME ?= $(notdir $(CURDIR))
MAIN_PACKAGE ?= .
BUILD_DIR ?= bin
BINARY ?= $(BUILD_DIR)/$(APP_NAME)
GOFLAGS ?=
GO_TEST_FLAGS ?=
LDFLAGS ?=
ARGS ?=

.DEFAULT_GOAL := help

.PHONY: all help deps tidy fmt vet test test-cover check build run clean

all: build ## Build the application binary

help: ## Show available make targets
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_.-]+:.*## / {printf "%-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download Go module dependencies
	$(GO) mod download

tidy: ## Tidy module dependencies
	$(GO) mod tidy

fmt: ## Format Go source files
	$(GO) fmt ./...

vet: ## Run static analysis with go vet
	$(GO) vet ./...

test: ## Run the test suite
	$(GO) test $(GO_TEST_FLAGS) ./...

test-cover: ## Run tests with coverage enabled
	$(GO) test $(GO_TEST_FLAGS) -cover ./...

check: fmt vet test ## Run local quality checks

build: ## Build the application binary
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) $(MAIN_PACKAGE)

run: ## Run the application locally
	$(GO) run $(GOFLAGS) $(MAIN_PACKAGE) $(ARGS)

clean: ## Remove generated build artifacts
	rm -rf $(BUILD_DIR)

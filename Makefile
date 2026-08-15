APP_NAME := stress-tester
BUILD_DIR := build
GO ?= go
GOFLAGS ?=

.DEFAULT_GOAL := help
.PHONY: help fmt test build run tidy update clean audit

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make <target>\n\nTargets:\n"} /^[a-zA-Z_-]+:.*##/ {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt: ## Format Go sources
	gofmt -w $$(find . -name "*.go" -not -path "./vendor/*")

test: ## Run tests and compile all packages
	$(GO) test $(GOFLAGS) ./...

build: ## Build the desktop application
	mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .

run: ## Run the application from source
	$(GO) run $(GOFLAGS) .

tidy: ## Synchronize go.mod and go.sum
	$(GO) mod tidy

update: ## Update Go dependencies
	$(GO) get -u ./...
	$(GO) mod tidy

clean: ## Remove build artifacts and logs
	rm -rf $(BUILD_DIR) stress_test.log

audit: fmt test ## Run the default local audit checks

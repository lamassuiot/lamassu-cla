# Developer entry points for the CLA Management Platform (F-01).
# Run `make help` for targets. Requires Node.js 24 (`nvm use`), npm, Go, and Docker for the Dev Container.

SHELL := bash
.SHELLFLAGS := -euo pipefail -c
.DEFAULT_GOAL := help

GO_DIR := services/api
GO_TOOLCHAIN := go1.27.1
GOLANGCI_LINT_VERSION := v2.14.0
TOOLS_BIN := $(CURDIR)/.tools/bin
GOLANGCI_LINT := $(TOOLS_BIN)/golangci-lint
LAMBDA_FUNCTIONS := api

export GOTOOLCHAIN := $(GO_TOOLCHAIN)

.PHONY: help
help: ## List targets.
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-14s %s\n", $$1, $$2}'

.PHONY: bootstrap
bootstrap: tools ## Install npm dependencies from the lockfile and pinned Go tools.
	npm ci
	cd $(GO_DIR) && go mod download

.PHONY: tools
tools: $(GOLANGCI_LINT) ## Install pinned Go tools into .tools/bin.

$(GOLANGCI_LINT):
	GOBIN=$(TOOLS_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: lint
lint: lint-web lint-go lint-repo ## Run all linters.

.PHONY: lint-web
lint-web: ## Biome lint and format check, then the TypeScript type check.
	npm run lint
	npm run typecheck

.PHONY: lint-go
lint-go: $(GOLANGCI_LINT)
	cd $(GO_DIR) && go vet ./...
	cd $(GO_DIR) && $(GOLANGCI_LINT) run ./...

.PHONY: lint-repo
lint-repo:
	scripts/check-private-urls.sh
	scripts/check-ts-directives.sh

.PHONY: test
test: test-web test-go test-scripts ## Run all tests.

.PHONY: test-web
test-web:
	npm test

.PHONY: test-go
test-go:
	cd $(GO_DIR) && go test -race ./...

.PHONY: test-scripts
test-scripts: $(GOLANGCI_LINT)
	scripts/check-private-urls.test.sh
	scripts/check-ts-directives.test.sh
	GOLANGCI_LINT=$(GOLANGCI_LINT) scripts/check-go-layering.test.sh

.PHONY: build
build: build-web build-go ## Build the portal and Lambda binaries.

.PHONY: build-web
build-web:
	npm run build

.PHONY: build-go
build-go:
	@for fn in $(LAMBDA_FUNCTIONS); do \
		echo "building $$fn"; \
		(cd $(GO_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -tags lambda.norpc -o dist/$$fn/bootstrap ./cmd/$$fn); \
	done

.PHONY: format
format: ## Format and apply safe Biome fixes to web sources; format Go sources.
	npm run format
	cd $(GO_DIR) && $(GOLANGCI_LINT) fmt ./...

.PHONY: check
check: lint test build ## Run lint, test, and build, as CI does.

.PHONY: clean
clean: ## Remove build output.
	rm -rf apps/web/dist $(GO_DIR)/dist

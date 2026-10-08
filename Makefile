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
GOVULNCHECK_VERSION := v1.8.0
GOVULNCHECK := $(TOOLS_BIN)/govulncheck
LAMBDA_FUNCTIONS := api

# Terraform (F-03). Terraform and TFLint come from the Dev Container (versions in .devcontainer).
TF_DIRS := $(sort $(wildcard infra/modules/*)) infra/bootstrap $(sort $(wildcard infra/environments/*))
TF_PLUGIN_CACHE := $(CURDIR)/.tools/terraform-plugin-cache

# Checkov (A-19, ADR-0012), installed only from the hash-locked requirements with Python 3.12.
CHECKOV_PYTHON ?= python3
CHECKOV_LOCK := scripts/checkov/requirements.txt
CHECKOV_VENV := $(CURDIR)/.tools/checkov
CHECKOV_BIN := $(CHECKOV_VENV)/bin/checkov
CHECKOV_HOME := $(CURDIR)/.tools/checkov-home
# An empty environment keeps AWS and Prisma Cloud credentials away from Checkov.
CHECKOV_ENV := env -i PATH=$(CHECKOV_VENV)/bin:/usr/bin:/bin HOME=$(CHECKOV_HOME) LC_ALL=C.UTF-8
# Offline, findings carry no severity, so any failed check fails the scan (no --soft-fail or
# --hard-fail-on). Exceptions are inline `checkov:skip=<ID>:<reason>` comments.
CHECKOV_FLAGS := --compact --framework terraform --download-external-modules false --skip-download --output cli
PIP_TOOLS_VERSION := 7.5.1

export GOTOOLCHAIN := $(GO_TOOLCHAIN)

.PHONY: help
help: ## List targets.
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.PHONY: bootstrap
bootstrap: tools ## Install npm dependencies from the lockfile and pinned Go tools.
	npm ci
	cd $(GO_DIR) && go mod download

.PHONY: tools
tools: $(GOLANGCI_LINT) $(GOVULNCHECK) ## Install pinned Go tools into .tools/bin.

$(GOLANGCI_LINT):
	GOBIN=$(TOOLS_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

$(GOVULNCHECK):
	GOBIN=$(TOOLS_BIN) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

.PHONY: lint
lint: lint-web lint-openapi lint-go lint-repo ## Run all linters.

.PHONY: lint-web
lint-web: ## Biome lint and format check, then the TypeScript type check.
	npm run lint
	npm run typecheck

.PHONY: lint-openapi
lint-openapi: ## Lint the OpenAPI contract with Redocly (redocly.yaml).
	npm run lint:openapi

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

.PHONY: vuln-go
vuln-go: $(GOVULNCHECK) ## Scan the Go module and toolchain for known vulnerabilities (needs network).
	cd $(GO_DIR) && $(GOVULNCHECK) ./...

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

.PHONY: check-terraform
check-terraform: fmt-terraform validate-terraform test-terraform lint-terraform scan-terraform ## Terraform fmt, validate, tests, TFLint, and Checkov (no AWS access).

.PHONY: fmt-terraform
fmt-terraform: ## Check Terraform formatting.
	terraform fmt -check -recursive -diff infra

.PHONY: validate-terraform
validate-terraform: ## Validate every Terraform module and root without a backend.
	@mkdir -p $(TF_PLUGIN_CACHE)
	@for dir in $(TF_DIRS); do \
		echo "validate $$dir"; \
		lockfile=readonly; [[ $$dir == infra/modules/* ]] && lockfile=; \
		TF_PLUGIN_CACHE_DIR=$(TF_PLUGIN_CACHE) terraform -chdir=$$dir init -backend=false -input=false -no-color $${lockfile:+-lockfile=$$lockfile} >/dev/null; \
		terraform -chdir=$$dir validate -no-color; \
	done

.PHONY: test-terraform
test-terraform: validate-terraform ## Run plan-only Terraform tests with mocked providers.
	@for dir in $(TF_DIRS); do \
		echo "test $$dir"; \
		terraform -chdir=$$dir test -no-color; \
	done

.PHONY: lint-terraform
lint-terraform: ## Run TFLint with the pinned AWS ruleset on infra/.
	tflint --init --config $(CURDIR)/.tflint.hcl
	cd infra && tflint --recursive --config $(CURDIR)/.tflint.hcl --format compact

.PHONY: scan-terraform
scan-terraform: test-terraform-scan ## Scan infra/ with Checkov; any failed check fails (no AWS access).
	$(CHECKOV_ENV) $(CHECKOV_BIN) $(CHECKOV_FLAGS) --directory infra

.PHONY: test-terraform-scan
test-terraform-scan: $(CHECKOV_BIN) ## Check that Checkov fails on the insecure fixtures in scripts/testdata/checkov.
	@mkdir -p $(CHECKOV_HOME)
	$(CHECKOV_ENV) CHECKOV_BIN=$(CHECKOV_BIN) CHECKOV_FLAGS="$(CHECKOV_FLAGS)" scripts/check-terraform-scan.test.sh

$(CHECKOV_BIN): $(CHECKOV_LOCK)
	@$(CHECKOV_PYTHON) -c 'import sys; sys.exit(sys.version_info[:2] != (3, 12))' || { echo "Checkov needs Python 3.12 as $(CHECKOV_PYTHON); use the Dev Container"; exit 1; }
	rm -rf $(CHECKOV_VENV)
	$(CHECKOV_PYTHON) -m venv $(CHECKOV_VENV)
	$(CHECKOV_VENV)/bin/pip install --quiet --disable-pip-version-check --no-deps --require-hashes --only-binary=:all: -r $(CHECKOV_LOCK)
	$(CHECKOV_VENV)/bin/pip check

.PHONY: lock-checkov
lock-checkov: ## Regenerate the hash-locked Checkov requirements (Python 3.12, network).
	rm -rf $(CURDIR)/.tools/pip-tools
	$(CHECKOV_PYTHON) -m venv $(CURDIR)/.tools/pip-tools
	$(CURDIR)/.tools/pip-tools/bin/pip install --quiet --disable-pip-version-check pip-tools==$(PIP_TOOLS_VERSION)
	cd scripts/checkov && CUSTOM_COMPILE_COMMAND="pip-compile --allow-unsafe --generate-hashes --output-file=requirements.txt --strip-extras requirements.in" $(CURDIR)/.tools/pip-tools/bin/pip-compile --quiet --generate-hashes --allow-unsafe --strip-extras --output-file requirements.txt requirements.in

.PHONY: clean
clean: ## Remove build output.
	rm -rf apps/web/dist $(GO_DIR)/dist

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
TRIVY_VERSION := 0.75.0
TRIVY_SHA256_amd64 := c6e65abddb348e25f10549df887045629cf28cc72453cd1c63acb717316b3f3f
TRIVY_SHA256_arm64 := a1ee9f6ffb7d112b64ff726a2a0717c21175c1114361391f4a132956751a13b3
TRIVY := $(TOOLS_BIN)/trivy-$(TRIVY_VERSION)
TF_PLUGIN_CACHE := $(CURDIR)/.tools/terraform-plugin-cache

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
check-terraform: fmt-terraform validate-terraform test-terraform lint-terraform scan-terraform ## Terraform fmt, validate, tests, TFLint, and Trivy (no AWS access).

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
lint-terraform: ## Run TFLint with the pinned AWS ruleset.
	tflint --init --config $(CURDIR)/.tflint.hcl
	tflint --recursive --config $(CURDIR)/.tflint.hcl --format compact

.PHONY: scan-terraform
scan-terraform: $(TRIVY) ## Scan Terraform configuration with Trivy (fails on high and critical findings).
	$(TRIVY) config --quiet --disable-telemetry --skip-version-check --skip-dirs '**/.terraform' --severity HIGH,CRITICAL --exit-code 1 infra

$(TRIVY):
	@arch=$$(uname -m); case $$arch in x86_64) arch=amd64; asset=64bit;; aarch64|arm64) arch=arm64; asset=ARM64;; *) echo "unsupported architecture $$arch"; exit 1;; esac; \
	[[ $$(uname -s) == Linux ]] || { echo "Trivy install supports Linux only; use the Dev Container"; exit 1; }; \
	sha=$$( [[ $$arch == amd64 ]] && echo $(TRIVY_SHA256_amd64) || echo $(TRIVY_SHA256_arm64) ); \
	tmp=$$(mktemp -d); trap 'rm -rf "$$tmp"' EXIT; \
	curl -fsSLo "$$tmp/trivy.tgz" "https://github.com/aquasecurity/trivy/releases/download/v$(TRIVY_VERSION)/trivy_$(TRIVY_VERSION)_Linux-$$asset.tar.gz"; \
	echo "$$sha  $$tmp/trivy.tgz" | sha256sum --check --strict; \
	tar -xzf "$$tmp/trivy.tgz" -C "$$tmp" trivy; \
	mkdir -p $(TOOLS_BIN); install -m 0755 "$$tmp/trivy" $@

.PHONY: clean
clean: ## Remove build output.
	rm -rf apps/web/dist $(GO_DIR)/dist

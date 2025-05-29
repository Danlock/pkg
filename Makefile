#! /usr/bin/make

SHELL = /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
MAKEFLAGS += --warn-undefined-variables
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

GITHASH = $(shell git describe --dirty --always)
GITCOMMITNO = $(shell git rev-list --all --count)
SHORTBUILDTAG = v0.0.$(GITCOMMITNO)-$(GITHASH)
BUILDTIME = $(shell date -u --rfc-3339=seconds)
BUILDINFO = Build Time:$(BUILDTIME)

.DEFAULT_GOAL := help
.PHONY: help
help:  ## Show this help
	@grep -E '^([a-zA-Z_-]+):.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "%-20s %s\n", $$1, $$2}'

.PHONY: deps
deps: ## Update dependencies
	go mod tidy
	go mod vendor

LDFLAGS = -X 'main.buildTag=$(SHORTBUILDTAG)' -X 'main.buildInfo=$(BUILDINFO)'
BINNAME = changeme
.PHONY: build
build: ## Build binary
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/$(BINNAME) ./cmd/$(BINNAME)

.PHONY: run
run: ## Run binary directly
	CGO_ENABLED=0 go run -ldflags "$(LDFLAGS)" ./...

.PHONY: version
version: ## Print current version
	@echo $(SHORTBUILDTAG)

COVERAGE_DIR ?= $(ROOT_DIR)/bin/coverage
TEST_FLAGS ?= -failfast -count=1 -covermode=atomic
.PHONY: test
test: ## Run all tests
	@rm -rf $(COVERAGE_DIR)/* || true
	@mkdir -p $(COVERAGE_DIR) || true
	@go test $(TEST_FLAGS) -coverprofile=$(COVERAGE_DIR)/cov.out ./...
	@go tool cover -func=$(COVERAGE_DIR)/cov.out | tail -n 1

.PHONY: coverage-html
coverage-html: ## Generate HTML coverage report
	@rm $(COVERAGE_DIR).html || true
	@go tool cover -html=$(COVERAGE_DIR)/cov.out -o $(COVERAGE_DIR).html

.PHONY: coverage-browser
coverage-browser: ## Open HTML coverage report in browser
	@go tool cover -html=$(COVERAGE_DIR)/cov.out

.PHONY: update-readme-badge
update-readme-badge: ## Update coverage badge within README.md
	@go tool cover -func=$(COVERAGE_DIR)/cov.out -o=$(COVERAGE_DIR)/cov.badge
	@go run github.com/AlexBeauchemin/gobadge@v0.3.0 -filename=$(COVERAGE_DIR)/cov.badge

PKGNAME = github.com/danlock/pkg
# pkg.go.dev documentation can be updated via go get updating the google proxy from another package
.PHONY: update-godocs
update-godocs: ## Update pkg.go.dev documentation
	@cd ../rmq; \
	GOPROXY=https://proxy.golang.org go get -u $(PKGNAME); \
	go mod tidy

.PHONY: release
release: ## Tag and push current commit as a release
	@$(MAKE) deps
ifeq ($(findstring dirty,$(SHORTBUILDTAG)),dirty)
	@echo "Version $(SHORTBUILDTAG) is filthy, commit to clean it" && exit 1
endif
	@read -t 10 -p "$(SHORTBUILDTAG) will be the new released version. Hit enter to proceed, CTRL-C to cancel."
	@git tag $(SHORTBUILDTAG)
	@git push origin $(SHORTBUILDTAG)

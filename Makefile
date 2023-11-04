#! /usr/bin/make
SHELL = /bin/bash
BUILDTIME = $(shell date -u --rfc-3339=seconds)
GITHASH = $(shell git describe --dirty --always --tags)
GITCOMMITNO = $(shell git rev-list --all --count)
SHORTBUILDTAG = $(GITCOMMITNO).$(GITHASH)
BUILDINFO = Build Time:$(BUILDTIME)
LDFLAGS = -X 'main.buildTag=$(SHORTBUILDTAG)' -X 'main.buildInfo=$(BUILDINFO)'
BINNAME = changeme
PKGNAME = github.com/danlock/pkg

.PHONY: build

depend: deps
deps:
	go mod tidy
	go mod vendor

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/$(BINNAME) ./cmd/$(BINNAME)

docker-build:
	docker build -t $(BINNAME) .

run:
	CGO_ENABLED=0 go run -ldflags "$(LDFLAGS)" ./...

version:
	@echo $(SHORTBUILDTAG)

coverage:
	@go test -failfast -covermode=count -coverprofile=$(COVERAGE_PATH)

coverage-html:
	@rm $(COVERAGE_PATH) || true
	@$(MAKE) coverage
	@rm $(COVERAGE_PATH).html || true
	@go tool cover -html=$(COVERAGE_PATH) -o $(COVERAGE_PATH).html

coverage-browser:
	@rm $(COVERAGE_PATH) || true
	@$(MAKE) coverage
	@go tool cover -html=$(COVERAGE_PATH)

update-readme-badge:
	@go tool cover -func=$(COVERAGE_PATH) -o=$(COVERAGE_PATH).badge
	@go run github.com/AlexBeauchemin/gobadge@v0.3.0 -filename=$(COVERAGE_PATH).badge

# pkg.go.dev documentation is updated via go get updating the google proxy from another package
update-godocs:
	@cd ../rmq; \
	GOPROXY=https://proxy.golang.org go get -u $(PKGNAME); \
	go mod tidy

release:
	@$(MAKE) deps
ifeq ($(findstring dirty,$(SHORTBUILDTAG)),dirty)
	@echo "Version $(SHORTBUILDTAG) is filthy, commit to clean it" && exit 1
endif
	@read -t 5 -p "$(SHORTBUILDTAG) will be the new released version. Hit enter to proceed, CTRL-C to cancel."
	@git tag $(SHORTBUILDTAG)
	@git push origin $(SHORTBUILDTAG)

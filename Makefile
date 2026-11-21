OKF_BIN ?= $(shell command -v okf 2>/dev/null || echo okf)
BUNDLE  := knowledge
GOFMT   := $(shell go env GOROOT)/bin/gofmt

export CGO_ENABLED := 0

.PHONY: all build test validate search fmt lint godoc help generate css assets-fresh

all: validate test lint

build:
	go build -o bin/goisekai2 ./cmd/goisekai2

## generate: compile templ components and the Tailwind stylesheet
TEMPL_BIN ?= $(shell command -v templ 2>/dev/null || echo $(shell go env GOPATH)/bin/templ)
TEMPL_VERSION ?= v0.3.1118
TAILWIND_VERSION ?= 4.3.5
CSS_IN := internal/http/assets/input.css
CSS_OUT := internal/http/assets/app.css

generate: css
	@"$(TEMPL_BIN)" generate

css:
	@if command -v tailwindcss >/dev/null 2>&1; then \
		tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --minify; \
	else \
		npx --yes --package @tailwindcss/cli@$(TAILWIND_VERSION) tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --minify; \
	fi

## test: compile, vet and run the unit tests with cgo disabled
test:
	go build ./...
	go vet ./...
	go test ./...

## lint: gofmt the tree and enforce the godoc rule
lint: godoc sql-boundary assets-fresh
	@out="$$($(GOFMT) -l .)"; \
	if [ -n "$$out" ]; then echo "not gofmt-clean:"; echo "$$out"; exit 1; fi

## godoc: package comments and name-first doc comments on exported symbols
godoc:
	@go run scripts/godoc-check.go

## sql-boundary: raw SQL may live only in internal/store/sql/queries.go
sql-boundary:
	@go run scripts/sql-boundary-check.go

## assets-fresh: regeneration must be a no-op in a clean tree
assets-fresh:
	@out="$$(git diff --name-only -- internal/http/assets internal/http/components internal/http/*.templ.go 2>/dev/null)"; \
	if [ -n "$$out" ]; then echo "generated assets are stale:"; echo "$$out"; exit 1; fi

## validate: strict OKF v0.2 validation on the knowledge/ bundle
validate:
	@$(OKF_BIN) validate $(BUNDLE) --strict --drift

fmt:
	$(GOFMT) -l -w .

search:
	@$(OKF_BIN) search "$(q)" $(BUNDLE)

help:
	@echo "build | generate | test | lint | godoc | validate | search q=<term> | fmt"

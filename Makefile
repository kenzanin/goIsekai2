OKF_BIN ?= $(shell command -v okf 2>/dev/null || echo okf)
BUNDLE  := knowledge
GOFMT   := $(shell go env GOROOT)/bin/gofmt

export CGO_ENABLED := 0

.PHONY: all build test validate search fmt lint godoc help

all: validate test lint

build:
	go build -o bin/goisekai2 ./cmd/goisekai2

## test: compile, vet and run the unit tests with cgo disabled
test:
	go build ./...
	go vet ./...
	go test ./...

## lint: gofmt the tree and enforce the godoc rule
lint: godoc
	@out="$$($(GOFMT) -l .)"; \
	if [ -n "$$out" ]; then echo "not gofmt-clean:"; echo "$$out"; exit 1; fi

## godoc: package comments and name-first doc comments on exported symbols
godoc:
	@go run scripts/godoc-check.go

## validate: strict OKF v0.2 validation on the knowledge/ bundle
validate:
	@$(OKF_BIN) validate $(BUNDLE) --strict --drift

fmt:
	$(GOFMT) -l -w .

search:
	@$(OKF_BIN) search "$(q)" $(BUNDLE)

help:
	@echo "build | test | lint | godoc | validate | search q=<term> | fmt"

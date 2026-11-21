OKF_BIN ?= $(shell command -v okf 2>/dev/null || echo okf)
BUNDLE  := knowledge

export CGO_ENABLED := 0

.PHONY: all build test validate search fmt help

all: validate test

build:
	go build -o bin/goisekai2 ./cmd/goisekai2

test:
	go vet ./...
	go test ./...

## validate: strict OKF v0.2 validation on the knowledge/ bundle
validate:
	@$(OKF_BIN) validate $(BUNDLE) --strict --drift

fmt:
	gofmt -l -w .

search:
	@$(OKF_BIN) search "$(q)" $(BUNDLE)

help:
	@echo "build | test | validate | search q=<term> | fmt"

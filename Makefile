## ─── Config ─────────────────────────────────────────────────────────────────
BINARY     := mind-forge
MAIN       := ./main.go
GOBIN      := $(shell go env GOPATH)/bin
TEMPL      := $(shell which templ 2>/dev/null || echo "$(GOBIN)/templ")
TAILWIND   := $(shell which tailwindcss 2>/dev/null || echo "$(HOME)/.local/bin/tailwindcss")
TEMPL_OUT  := ./internal/templates
CSS_IN     := ./styles/input.css
CSS_OUT    := ./static/css/output.css

## ─── Default ─────────────────────────────────────────────────────────────────
.DEFAULT_GOAL := dev

## ─── Dev (watch everything) ──────────────────────────────────────────────────
.PHONY: dev
dev:
	@echo "→ starting dev watchers (templ + tailwind + go run)"
	@$(MAKE) -j3 templ/watch css/watch run/watch

.PHONY: run/watch
run/watch:
	go run github.com/air-verse/air@latest

## ─── Build ───────────────────────────────────────────────────────────────────
.PHONY: build
build: templ/gen css/build
	go build -o $(BINARY) $(MAIN)

## ─── templ ───────────────────────────────────────────────────────────────────
.PHONY: templ/gen
templ/gen:
	$(TEMPL) generate

.PHONY: templ/watch
templ/watch:
	$(TEMPL) generate --watch

## ─── Tailwind ────────────────────────────────────────────────────────────────
.PHONY: css/build
css/build:
	$(TAILWIND) -i $(CSS_IN) -o $(CSS_OUT) --minify

.PHONY: css/watch
css/watch:
	$(TAILWIND) -i $(CSS_IN) -o $(CSS_OUT) --watch

## ─── Tests ───────────────────────────────────────────────────────────────────
.PHONY: test
test:
	go test -race ./...

.PHONY: test/cover
test/cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## ─── Lint ────────────────────────────────────────────────────────────────────
.PHONY: lint
lint:
	golangci-lint run ./...
	go vet ./...

## ─── Deps ────────────────────────────────────────────────────────────────────
.PHONY: deps
deps:
	npm install
	go mod tidy
	go mod download

## ─── Clean ───────────────────────────────────────────────────────────────────
.PHONY: clean
clean:
	rm -f $(BINARY) $(CSS_OUT)
	find $(TEMPL_OUT) -name "*_templ.go" -delete

## ─── Help ────────────────────────────────────────────────────────────────────
.PHONY: help
help:
	@grep -E '^[a-zA-Z/_-]+:.*?##' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  %-20s %s\n",$$1,$$2}'

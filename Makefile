# go-concurrency — Makefile
# Common developer tasks. Run `make help` to list them.

GO      ?= go
PKGS    := ./...
BIN     := demo
ARGS    ?=

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: ## Run a demo, e.g. make run ARGS="waitgroup --tasks 8"
	$(GO) run ./cmd/demo $(ARGS)

.PHONY: build
build: ## Build the demo binary into ./bin
	@mkdir -p bin
	$(GO) build -o bin/$(BIN) ./cmd/demo

.PHONY: test
test: ## Run all tests with the race detector and coverage
	$(GO) test -race -cover $(PKGS)

.PHONY: bench
bench: ## Run all benchmarks
	$(GO) test -run '^$$' -bench . -benchmem $(PKGS)

.PHONY: vet
vet: ## Run go vet
	$(GO) vet $(PKGS)

.PHONY: lint
lint: ## Run golangci-lint (must be installed)
	golangci-lint run

.PHONY: fmt
fmt: ## Format the code
	gofmt -w .

.PHONY: coverage
coverage: ## Generate an HTML coverage report at coverage.html
	$(GO) test -race -coverprofile=coverage.out $(PKGS)
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "open coverage.html"

.PHONY: docker-build
docker-build: ## Build the Docker image
	docker build -t go-concurrency .

.PHONY: docker-run
docker-run: ## Run a demo in Docker, e.g. make docker-run ARGS="fanin"
	docker run --rm go-concurrency $(ARGS)

.PHONY: clean
clean: ## Remove build and coverage artifacts
	rm -rf bin coverage.out coverage.html

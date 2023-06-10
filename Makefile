.DEFAULT_GOAL := help

BIN_NAME := build/abihelper
PKG := abihelper
VERSION := 1.0.0

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: deps
deps: ## Install dependencies
ifeq ($(OS),Windows_NT) # Windows
	choco install golang sqlite golangci-lint redis
else
	UNAME_S := $(shell uname -s 2>/dev/null || echo "unknown")
	ifeq ($(UNAME_S),Linux) # Linux
		sudo apt-get update && sudo apt-get install -y golang sqlite3 golangci-lint redis-server
	endif
	ifeq ($(UNAME_S),Darwin) # macOS
		brew install go sqlite golangci-lint redis
	endif
endif

.PHONY: lint
lint: ## Lint the Go code using golangci-lint
	golangci-lint run

.PHONY: build
build: ## Build the binary for the current OS/Arch
ifeq ($(OS),Windows_NT) # Windows
	@make build-windows
else
	UNAME_S := $(shell uname -s 2>/dev/null || echo "unknown")
	ifeq ($(UNAME_S),Linux) # Linux
		@make build-linux
	endif
	ifeq ($(UNAME_S),Darwin) # macOS
		@make build-macos
	endif
endif

.PHONY: build-linux
build-linux: ## Build the binary for Linux
	GOOS=linux GOARCH=amd64 go build -o ./$(BIN_NAME) -ldflags "-X main.Version=$(VERSION)" .

.PHONY: build-macos
build-macos: ## Build the binary for MacOS
	GOOS=darwin GOARCH=amd64 go build -o ./$(BIN_NAME) -ldflags "-X main.Version=$(VERSION)" .

.PHONY: build-windows
build-windows: ## Build the binary for Windows
	GOOS=windows GOARCH=amd64 go build -o ./$(BIN_NAME).exe -ldflags "-X main.Version=$(VERSION)" .

.PHONY: test
test: ## Run tests
	go test -v -cover ./...

.PHONY: benchmark
benchmark: ## Run benchmarks
	go test -v -bench . -benchmem ./... > benchmark.txt

.PHONY: clean
clean: ## Clean build files
ifeq ($(OS),Windows_NT) # Windows
	del /Q $(BIN_NAME).exe
else
	rm -f $(BIN_NAME)
endif

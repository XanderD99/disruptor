-include .makerc
SHELL = bash
PACKAGE ?= disruptor

# Build server placeholders (BUILD_NUMBER and BRANCH_NAME are overwritten by Jenkins)
BUILD_NUMBER ?= 1
BRANCH_NAME ?= local
BRANCH_NAME_CLEAN = $(shell echo '$(BRANCH_NAME)' | tr "[:upper:]" "[:lower:]" | tr "/" "-")
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "0.1.0")

# Static
GIT_COMMIT ?= $(shell git rev-parse HEAD 2> /dev/null)
GIT_AUTHORS ?= $(shell git log --format='%aN' | sort -u | awk -vORS=, '{ print }' | sed 's/,$$//')

# Output purposes
OUTPUT_DIR = $(CURDIR)/output
BIN_OUTPUT_DIR = $(OUTPUT_DIR)/bin
TEST_OUTPUT_DIR = $(OUTPUT_DIR)/test
DIRS=$(BIN_OUTPUT_DIR) $(TEST_OUTPUT_DIR)
$(shell mkdir -p $(DIRS))

# Build flags
MAIN_PACKAGE = ./cmd/disruptor
LDFLAGS ?= "-X 'main.version=$(VERSION)'"

BUILD_FLAGS ?= $(EXTRA_BUILD_FLAGS) $(MAIN_PACKAGE)

LINT_FLAGS ?= -c ./ci/config/.golangci.yaml $(EXTRA_LINT_FLAGS)
TEST_FLAGS ?= "-tags=unit"
GO111MODULE=on
CGO_ENABLED=1

# Docker
DOCKER_BUILD_ARGS ?=--build-arg ARG_GIT_COMMIT=$(GIT_COMMIT) --build-arg ARG_VERSION=$(VERSION) --build-arg ARG_AUTHORS="$(GIT_AUTHORS)"
DOCKER_FILE_PATH ?= ./ci/Dockerfile

.NOTPARALLEL: ; # wait for this target to finish
.EXPORT_ALL_VARIABLES: ; # send all vars to shell
.PHONY: all clean version build docs test scripts api cmd configs examples deps

deps: ## Add dependencies for your project

version: ## Return version
ifeq ($(BRANCH_NAME), main)
	@echo `cat version 2> /dev/null || echo 0.0.1`
else
	@echo $(VERSION)
endif

help: ## Show Help
	@echo "Usage:"
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort |\
	awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

lint: ## Lint
	golangci-lint run $(LINT_FLAGS)

build: ## Build the app
	go build -o $(BIN_OUTPUT_DIR)/$(PACKAGE)$(PACKAGE_EXTENSION) --ldflags=$(LDFLAGS) $(BUILD_FLAGS)

run: ## Run the app
	$(BIN_OUTPUT_DIR)/$(PACKAGE)$(PACKAGE_EXTENSION)

test: ## Run tests
	go test $(TEST_FLAGS) ./... $(BUILD_FLAGS)

docker: ## Build: docker
	docker build --target final $(DOCKER_BUILD_ARGS) -t $(PACKAGE):$(VERSION) -f $(DOCKER_FILE_PATH) .

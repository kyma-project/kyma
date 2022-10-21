# Inspired by:
# https://betterprogramming.pub/my-ultimate-makefile-for-golang-projects-fcc8ca20c9bb
SHELL=/bin/bash

GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_NAME=example
VERSION?=0.0.0

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all test test-k3d

all: help

## Test:

test: ## Run the tests of the project
	$(GOTEST) -run TestApiGateway . -v -race

test-custom-domain:
	$(GOTEST) -run "^TestCustomDomain$$" . -v -race

test-k3d: ## Prepare k3d and run tests
	export KYMA_DOMAIN="local.kyma.dev"
	export TEST_HYDRA_ADDRESS="https://oauth2.local.kyma.dev"
	export TEST_REQUEST_TIMEOUT="120"
	export TEST_REQUEST_DELAY="10"
	export TEST_DOMAIN="local.kyma.dev" 
	export TEST_CLIENT_TIMEOUT=30s
	export TEST_CONCURENCY="8"
	export EXPORT_RESULT="true"

	./prepare_local.sh

	make test

## Prerequisites:
setup: ## Setup local installation with the same configuration as pipeline
	./prepare_local.sh
## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)
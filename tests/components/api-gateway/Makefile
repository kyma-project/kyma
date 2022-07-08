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

.PHONY: all test

all: help

## Test:

test: ## Run the tests of the project
ifeq ($(EXPORT_RESULT), true)
	$(GOTEST) -v -race . --godog.format=pretty,junit:junit-report.xml,cucumber:cucumber-report.json
else
	$(GOTEST) -v -race .
endif

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
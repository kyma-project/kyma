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

.PHONY: all test provision-k3d kyma-deploy test-k3d

all: help

## Test:

test: ## Run the tests of the project
ifeq ($(EXPORT_RESULT), true)
	$(GOTEST) -v -race ./... --godog.format=pretty,junit:junit-report.xml,cucumber:cucumber-report.json
else
	$(GOTEST) -v -race ./...
endif

test-k3d: provision-k3d kyma-istio-deploy test ## Run tests on k3d cluster

provision-k3d: ## Provision k3d cluster
	kyma provision k3d --port 8080:80@loadbalancer --ci

kyma-istio-deploy: ## Deploy Kyma Istio using profile is set in KYMA_PROFILE env variable
ifndef KYMA_PROFILE
$(error KYMA_PROFILE is not set)
endif
	kyma deploy -s main --component cluster-essentials --component istio --profile ${KYMA_PROFILE} --ci

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
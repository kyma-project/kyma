APP_NAME = application-gateway-tests
APP_PATH = tests/application-gateway-tests
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20200423-1d9d6590
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

test:
	@echo "Tests are not to be run at this stage."

resolve-local:
	GO111MODULE=on go mod vendor -v

ensure-local:
	@echo "Go modules present in component - omitting."

dep-status:
	@echo "Go modules present in component - omitting."

dep-status-local:
	@echo "Go modules present in component - omitting."

verify-local:
	GO111MODULE=on go mod verify

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

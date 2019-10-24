APP_NAME = application-gateway-tests
APP_PATH = tests/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.11
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

.PHONY: verify format release vet

# TEST_EXECUTOR
APP_NAME = application-gateway-tests-executor
APP_PATH = tests/$(APP_NAME)

verify:: vet

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

# DEP ENSURE
# DEP STATUS
# GO IMPORTS
# GO VET
# here# here# here
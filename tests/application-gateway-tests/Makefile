APP_NAME = application-gateway-tests
APP_PATH = tests/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/go-dep.mk

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

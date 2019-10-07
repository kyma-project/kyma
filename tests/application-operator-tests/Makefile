APP_NAME = application-operator-tests
APP_PATH = tests/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/go-dep.mk

VERIFY_IGNORE := /vendor\|/mocks

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

verify:: vet

# DEP ENSURE
# DEP STATUS
# GO FMT
# GO VET

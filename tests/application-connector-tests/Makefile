APP_NAME = application-connector-tests
APP_PATH = tests/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.11
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/go-dep.mk

VERIFY_IGNORE := /vendor\|/mocks
NAMESPACE := kyma-integration
CENTRAL := false
SKIP_SSL_VERIFY := true

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

verify:: vet

# DEP ENSURE
# DEP STATUS
# GO FMT
# GO VET
# here
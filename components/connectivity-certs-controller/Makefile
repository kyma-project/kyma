APP_NAME = connectivity-certs-controller
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.11
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

verify:: vet

# DEP ENSURE
# GO BUILD
# DEP STATUS
# GO TEST
# Check if tests passed
# GO FMT
# GO VET
# here# here# here
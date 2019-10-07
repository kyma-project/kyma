APP_NAME = compass-runtime-agent
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/go-dep.mk

VERIFY_IGNORE := /vendor\|/mocks

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/compass-runtime-agent"

verify:: vet

# GET DEPENDENCIES
# GO BUILD
# GO TEST
# Check if tests passed
# GO FMT
# GO VET

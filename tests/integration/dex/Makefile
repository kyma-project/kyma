APP_NAME = dex-integration-tests
APP_PATH = tests/integration/dex
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
#IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
#TAG = $(DOCKER_TAG)
SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/scripts

include $(SCRIPTS_DIR)/go-dep.mk

# test target is overriden, so it won't run actual tests
test:
	@echo -n ""

verify:: vet

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

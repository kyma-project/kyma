APP_NAME = api-controller-integration-tests
APP_PATH = tests/integration/api-controller
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/scripts

include $(SCRIPTS_DIR)/go-dep.mk

# test target is overriden, so it won't run actual tests
test:
	@echo -n ""

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

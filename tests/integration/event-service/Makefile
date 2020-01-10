APP_NAME = event-service-integration-tests
APP_PATH = tests/integration/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219
SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

# test target is overriden, so it won't run actual tests
test: ;

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

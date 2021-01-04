APP_NAME = event-service-integration-tests
APP_PATH = tests/integration/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20201009-5d8104e0
SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

# test target is overriden, so it won't run actual tests
test-local: ;

release:
	$(MAKE) gomod-release

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

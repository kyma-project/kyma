APP_NAME = monitoring-integration-tests
APP_PATH = tests/integration/monitoring
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219
SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

# Override to use go modules
release:
	$(MAKE) gomod-release

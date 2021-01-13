APP_NAME = monitoring-integration-tests
APP_PATH = tests/integration/monitoring
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20201009-5d8104e0
SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

# Override to use go modules
release:
	$(MAKE) gomod-release

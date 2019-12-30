APP_NAME = istio-injection-integration-tests
APP_PATH = tests/integration/istio-injection
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

# test target is overriden, so it won't run actual tests
test:
	@echo -n ""

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

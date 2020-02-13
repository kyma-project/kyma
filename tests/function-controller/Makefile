APP_NAME = function-controller-test
APP_PATH = tests/function-controller
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20191011-51ed45a
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

test:
	@echo "Tests are not to be run at this stage."

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/function-controller"
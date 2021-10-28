APP_NAME = dex-integration-tests
APP_PATH = tests/integration/dex
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20201009-5d8104e0
SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

verify:: vet

release:
	$(MAKE) gomod-release

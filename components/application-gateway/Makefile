APP_NAME = application-gateway
APP_PATH = components/$(APP_NAME)
# BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20200423-1d9d6590
SCRIPTS_DIR = $(realpath $(shell pwd))

override MOUNT_TARGETS = ''

override ENTRYPOINT = cmd/applicationgateway/

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

release:
	$(MAKE) gomod-release-local



.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

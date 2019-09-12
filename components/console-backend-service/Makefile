APP_NAME = console-backend-service
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/pr/buildpack-golang-toolbox:PR-1469
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/go-dep.mk

VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

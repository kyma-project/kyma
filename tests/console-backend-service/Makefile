APP_NAME = console-backend-service-test
APP_PATH = tests/console-backend-service
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/go-dep.mk

VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg

verify:: errcheck

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

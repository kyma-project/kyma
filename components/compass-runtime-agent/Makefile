APP_NAME = compass-runtime-agent
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20190823-24e14d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

verify:: mod-verify

resolve:
	go mod vendor -v

dep-status:
	@echo "Go modules present in component - omitting."

mod-tidy:
	GO111MODULE=on go mod tidy -v

mod-verify:
	go mod verify 

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/compass-runtime-agent"

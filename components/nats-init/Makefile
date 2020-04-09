APP_NAME = nats-init
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/docker.mk

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"


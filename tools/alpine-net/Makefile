APP_NAME = alpine-net
APP_PATH = tools/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/bootstrap:v20190604-d08e7fe
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/docker.mk

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/helm-broker"

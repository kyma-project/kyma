APP_NAME = login-consent
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
include $(SCRIPTS_DIR)/generic-make-go.mk

ENTRYPOINT := ./main.go

verify:: vet

.PHONY: clean
clean:
	rm -rf bin

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/ory/login-consent"

APP_NAME = apiserver-proxy
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
include $(SCRIPTS_DIR)/generic-make-go.mk

ENTRYPOINT := ./cmd/proxy/main.go

.DEFAULT_GOAL=gomod-verify

release: gomod-release

.PHONY: clean
>>>>>>> 5e64c9c03... fix generic makefile and clean up the code
clean:
	rm -rf bin

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

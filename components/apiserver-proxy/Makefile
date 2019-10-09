APP_NAME = apiserver-proxy
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
override MAIN_PATH := ./cmd/proxy/main.go

SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
include $(SCRIPTS_DIR)/go-dep.mk

verify:: vet

.PHONY: clean
clean:
	rm -rf bin

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

APP_NAME = kyma-operator
APP_PATH = components/kyma-operator
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/go-dep.mk

MAIN_PATH := ./cmd/operator/main.go

verify:: vet

.PHONY: clean
clean:
	rm -f $(binary)

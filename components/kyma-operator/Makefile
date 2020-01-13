APP_NAME = kyma-operator
APP_PATH = components/kyma-operator
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.13
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

ENTRYPOINT := ./cmd/operator/main.go

verify:: vet

.PHONY: clean
clean:
	rm -f $(binary)

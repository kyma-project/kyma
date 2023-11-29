APP_NAME = event-publisher-proxy
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20220407-4da6c929
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

# fail on lint issues
override IGNORE_LINTING_ISSUES =
override ENTRYPOINT = cmd/main.go

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

release:
	$(MAKE) gomod-release-local

.PHONY: clean
clean: resolve_clean

resolve_clean:
	rm -rf vendor

build-local: test-local

.PHONY: generate
generate:
	go generate ./...

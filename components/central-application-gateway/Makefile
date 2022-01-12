APP_NAME = central-application-gateway
APP_PATH = components/$(APP_NAME)
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

override ENTRYPOINT = cmd/applicationgateway/

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

release:
	$(MAKE) gomod-release-local

resolve-local:
	GO111MODULE=on go mod vendor -v

test-local:
	GO111MODULE=on go test ./...

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

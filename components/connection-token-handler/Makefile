APP_NAME = connection-token-handler
APP_PATH = components/$(APP_NAME)
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

override ENTRYPOINT = cmd/manager/

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

release:
	$(MAKE) gomod-release-local

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

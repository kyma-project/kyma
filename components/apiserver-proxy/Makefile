APP_NAME = apiserver-proxy
APP_PATH = components/$(APP_NAME)
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

release:
	$(MAKE) gomod-release-local

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

APP_NAME = alpine-net
APP_PATH = tools/$(APP_NAME)
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/docker.mk

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/helm-broker"

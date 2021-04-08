APP_NAME = event-service
APP_PATH = components/$(APP_NAME)
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

# COMPONENT_DIR is a local path to component
include $(SCRIPTS_DIR)/generic-make-go.mk

MAIN_PATH = ./cmd/eventservice/eventgateway.go

release:
	$(MAKE) gomod-release-local

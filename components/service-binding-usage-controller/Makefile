APP_NAME = service-binding-usage-controller
APP_PATH = components/$(APP_NAME)
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

ENTRYPOINT:=cmd/controller/main.go
VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg

release:
	$(MAKE) gomod-release-local

APP_NAME = application-broker
APP_PATH = components/$(APP_NAME)
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

ENTRYPOINT:=cmd/broker/main.go
VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg

release:
	$(MAKE) gomod-release-local

build-image: pull-licenses
	docker build -t $(IMG_NAME) -f cmd/broker/Dockerfile .

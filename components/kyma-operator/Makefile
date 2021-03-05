APP_NAME = kyma-operator
APP_PATH = components/$(APP_NAME)
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
override ENTRYPOINT := ./cmd/operator/main.go

include $(SCRIPTS_DIR)/generic-make-go.mk

release:
	$(MAKE) gomod-release-local

.PHONY: clean
clean:
	rm -f $(binary)

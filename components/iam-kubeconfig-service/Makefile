APP_NAME = iam-kubeconfig-service
APP_PATH = components/iam-kubeconfig-service
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

ENTRYPOINT := ./cmd/generator/main.go

release:
	$(MAKE) gomod-release-local

.PHONY: clean
clean:
	rm -rf bin

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

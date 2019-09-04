APP_NAME = console-backend-service
APP_PATH = components/$(APP_NAME)
BUILDPACK = sjanota/go-buildpack:1.11

include go.mk

VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

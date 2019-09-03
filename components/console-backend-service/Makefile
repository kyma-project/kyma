APP_NAME = console-backend-service
APP_PATH = components/$(APP_NAME)
BUILDPACK = sjanota/go-buildpack:1.11

include go.mk

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

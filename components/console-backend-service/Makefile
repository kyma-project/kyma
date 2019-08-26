APP_NAME = console-backend-service
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.11

include go.mk

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

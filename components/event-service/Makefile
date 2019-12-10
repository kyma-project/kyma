APP_NAME = event-service
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

# COMPONENT_DIR is a local path to component
include $(SCRIPTS_DIR)/generic-make-go.mk

MAIN_PATH = ./cmd/eventservice/eventgateway.go

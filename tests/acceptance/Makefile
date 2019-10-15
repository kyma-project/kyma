APP_NAME = acceptance-tests
APP_PATH = tests/acceptance
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
MAIN_PATH=cmd/controller/main.go

include $(SCRIPTS_DIR)/go-dep.mk

VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg


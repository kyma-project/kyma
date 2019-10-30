APP_NAME = service-catalog-acceptance-tests
APP_PATH = tests/service-catalog
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk
ENTRYPOINT=cmd/controller/main.go

VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg

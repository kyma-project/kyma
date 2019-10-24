APP_NAME = apiserver-proxy-integration-tests
APP_PATH = tests/integration/apiserver-proxy
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1

SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/common/makefiles
include $(SCRIPTS_DIR)/generic-make-go.mk

verify:: vet


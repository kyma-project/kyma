APP_NAME = apiserver-proxy-integration-tests
APP_PATH = tests/integration/apiserver-proxy
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1

SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/scripts
include $(SCRIPTS_DIR)/go-dep.mk

verify:: vet


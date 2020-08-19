APP_NAME = external-solution-end-to-end-tests
APP_PATH = tests/end-to-end/external-solution-integration
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20191011-51ed45a
SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/common/makefiles

override ENTRYPOINT=cmd/runner/main.go

include $(SCRIPTS_DIR)/generic-make-go.mk


VERIFY_IGNORE := /vendor\|/mocks

test:
	@echo "Tests are not to be run at this stage."

clustertest:
	cat testdefinition.ko.yaml| KYMADOMAIN=$(shell kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.domainName}') envsubst | ko apply -n kyma-system -f -
	kyma test run core-test-external-solution --watch

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

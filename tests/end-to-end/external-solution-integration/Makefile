APP_NAME = external-solution-end-to-end-tests
APP_PATH = tests/end-to-end/external-solution-integration
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20191011-51ed45a
SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/common/makefiles

ifeq ($(TEST_DEBUG_SCENARIO),)
TEST_DEBUG_SCENARIO := e2e-event-mesh
endif

override ENTRYPOINT=cmd/runner/main.go

include $(SCRIPTS_DIR)/generic-make-go.mk


VERIFY_IGNORE := /vendor\|/mocks

test:
	@echo "Tests are not to be run at this stage."

clustertest:
	rm .ko.yaml || true
	ln -s .ko.curl.yaml .ko.yaml
	cat testdefinition.ko.yaml | KYMADOMAIN=$(shell kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.domainName}') envsubst | ko apply -n kyma-system -f -
	kyma test run core-test-external-solution --watch

debug-local:
	kubectl delete pod -n kyma-system -l app=core-test-external-solution --ignore-not-found
	cat debug-locally-delve.ko.yaml | TEST_DEBUG_SCENARIO=$(TEST_DEBUG_SCENARIO) KYMADOMAIN=$(shell kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.domainName}') envsubst |  ko apply --disable-optimizations -f -

DELVE_IMAGE = $(KO_DOCKER_REPO)/dlv
build-push-delve-image:
	rm .ko.yaml || true
	ln -s .ko.delve.yaml .ko.yaml
	docker build -f Dockerfile.delve -t $(DELVE_IMAGE) .
	docker push $(DELVE_IMAGE)
	echo "defaultBaseImage: $(DELVE_IMAGE)" > .ko.delve.yaml

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

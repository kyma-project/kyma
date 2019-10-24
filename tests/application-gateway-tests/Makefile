APP_NAME = application-gateway-tests
TEST_EXECUTOR_APP_NAME = application-gateway-tests-executor
APP_PATH = tests/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.11
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
TEST_EXECUTOR_IMAGE = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(TEST_EXECUTOR_APP_NAME)

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

verify:: vet

test: 
	@echo "Tests are not to be run at this stage."

build-image: pull-licenses
	docker build -t $(APP_NAME):latest . -f ./test/helmtest/Dockerfile --build-arg TEST_EXECUTOR_IMAGE=$(TEST_EXECUTOR_IMAGE):$(TAG)
	docker build -t $(TEST_EXECUTOR_APP_NAME):latest . -f ./test/executor/Dockerfile

push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker tag $(TEST_EXECUTOR_APP_NAME) $(TEST_EXECUTOR_IMAGE):$(TAG)
	docker push $(IMG):$(TAG)
	docker push $(TEST_EXECUTOR_IMAGE):$(TAG)

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

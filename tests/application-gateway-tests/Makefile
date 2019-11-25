APP_NAME = application-gateway-tests
TEST_EXECUTOR_APP_NAME = application-gateway-tests-executor
APP_PATH = tests/$(APP_NAME)

BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20191011-51ed45a
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

TEST_EXECUTOR_IMG_NAME = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(TEST_EXECUTOR_APP_NAME)

VERIFY_IGNORE := /vendor\|/mocks

test: 
	@echo "Tests are not to be run at this stage."

build-image: pull-licenses
	docker build -t $(APP_NAME):latest . -f ./test/helmtest/Dockerfile --build-arg TEST_EXECUTOR_IMAGE=$(TEST_EXECUTOR_IMG_NAME):$(TAG)
	docker build -t $(TEST_EXECUTOR_APP_NAME):latest . -f ./test/executor/Dockerfile

push-image:
	docker tag $(APP_NAME) $(IMG_NAME):$(TAG)
	docker tag $(TEST_EXECUTOR_APP_NAME) $(TEST_EXECUTOR_IMG_NAME):$(TAG)
	docker push $(IMG_NAME):$(TAG)
	docker push $(TEST_EXECUTOR_IMG_NAME):$(TAG)

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

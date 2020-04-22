APP_NAME = external-solution-end-to-end-tests
APP_PATH = tests/end-to-end/external-solution-integration
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20191011-51ed45a
SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/common/makefiles

override ENTRYPOINT=cmd/runner/main.go

include $(SCRIPTS_DIR)/generic-make-go.mk


VERIFY_IGNORE := /vendor\|/mocks

test:
	@echo "Tests are not to be run at this stage."

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

####### Fake Lambda
FAKELAMBDA_NAME=fake-lambda
FAKELAMBDA_IMG_NAME := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(FAKELAMBDA_NAME)
TAG := $(DOCKER_TAG)

verify:: build-fake-lambda-image push-fake-lambda-image

.PHONY: build-fake-lambda-image push-fake-lambda-image

build-fake-lambda-image:
	docker build -t $(FAKELAMBDA_NAME) -f ./cmd/fakelambda/Dockerfile .

push-fake-lambda-image:
	docker tag $(FAKELAMBDA_NAME) $(FAKELAMBDA_IMG_NAME):$(TAG)
	docker push $(FAKELAMBDA_IMG_NAME):$(TAG)

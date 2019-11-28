APP_NAME = api-gateway-integration-tests
APP_PATH = tests/integration/api-gateway
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../../..)/scripts

IMG_NAME := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG := $(DOCKER_TAG)

# include $(SCRIPTS_DIR)/go-dep.mk

# test target is overriden, so it won't run actual tests
test:
	@echo -n ""

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

release: build-image push-image

build-image: pull-licenses
	docker build -t $(IMG_NAME) .

push-image:
	docker tag $(IMG_NAME) $(IMG_NAME):$(TAG)
	docker push $(IMG_NAME):$(TAG)

pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif
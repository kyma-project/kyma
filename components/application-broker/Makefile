APP_NAME = application-broker
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20201009-5d8104e0
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

ENTRYPOINT:=cmd/broker/main.go
VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg

release:
	$(MAKE) gomod-release

build-image: pull-licenses
	docker build -t $(IMG_NAME) -f cmd/broker/Dockerfile .

push-image:
	docker tag $(IMG_NAME) $(IMG_NAME):$(TAG)
	docker push $(IMG_NAME):$(TAG)

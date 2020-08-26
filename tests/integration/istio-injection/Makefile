APP_NAME = istio-injection-integration-tests
APP_PATH = tests/integration/istio-injection
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
	docker build --no-cache -t $(IMG_NAME) .

push-image:
	docker tag $(IMG_NAME) $(IMG_NAME):$(TAG)
	docker push $(IMG_NAME):$(TAG)

pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

# Skip dep commands

ensure-local:
	@echo "Go modules present in component - omitting."

dep-status:
	@echo "Go modules present in component - omitting."

dep-status-local:
	@echo "Go modules present in component - omitting."

# go mod

resolve-local:
	GO111MODULE=on go mod vendor -v

mod-verify-local:
	GO111MODULE=on go mod verify

test-local:
	GO111MODULE=on go test -coverprofile=/tmp/artifacts/cover.out ./...
	@echo -n "Total coverage: "
	@go tool cover -func=/tmp/artifacts/cover.out | grep total | awk '{print $$3}'

$(eval $(call buildpack-mount,resolve))
$(eval $(call buildpack-mount,mod-verify))
$(eval $(call buildpack-mount,test))

.PHONY: clean
clean:
	rm -f $(binary)

IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)
LOCAL_DIR = $(shell pwd)
WORKSPACE_DIR = /workspace/go/src/github.com/kyma-project/kyma/$(APP_PATH)

RED = '\033[0;31m'
GREEN = '\033[0;32m'
INVERTED = '\033[7m'
NC = '\033[0m' # No Color

.PHONY: release resolve build build-image push-image pull-licenses

define buildpack
	@echo "=> Buildpack: $(1)"
	@docker run --rm -v $(LOCAL_DIR):$(WORKSPACE_DIR) -w $(WORKSPACE_DIR) $(BUILDPACK) $(1)
endef

release: resolve build build-image push-image

resolve:
	$(call buildpack,dep ensure -v -vendor-only)

test:
	$(call buildpack,go test ./...)

build-image: pull-licenses
	docker build -t $(IMG_NAME) .

push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

before-commit: build dep-status test imports fmt errcheck vet

build:
	$(call buildpack,env CGO_ENABLED=0 go build -o $(APP_NAME))

dep-status:
	$(call buildpack, dep status -v)

pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif
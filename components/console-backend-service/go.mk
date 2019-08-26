IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)
LOCAL_DIR = $(shell pwd)
WORKSPACE_DIR = /workspace/go/src/github.com/kyma-project/kyma/$(APP_PATH)

.PHONY: release resolve build build-image push-image pull-licenses

define buildpack
	@docker run --rm -v $(LOCAL_DIR):$(WORKSPACE_DIR) -w $(WORKSPACE_DIR) $(BUILDPACK) $(1)
endef

release: resolve build build-image push-image

resolve:
	$(call buildpack,dep ensure -v -vendor-only)

test:
	$(call buildpack,go test ./...)

build:
	$(call buildpack,./before-commit.sh)

build-image: pull-licenses
	docker build -t $(IMG_NAME) .

push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif
TAG = $(DOCKER_TAG)

# event-publish-service
APP_NAME_PUBLISH_KN = event-publish-service
IMG_PUBLISH_KN = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME_PUBLISH_KN)
BINARY_PUBLISH_KN = $(APP_NAME_PUBLISH_KN)

# subscription-controller
APP_NAME_SCK = subscription-controller
IMG_SCK = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME_SCK)
BINARY_SCK = $(APP_NAME_SCK)

# event-bus-nats-init
APP_NAME_NI = event-bus-nats-init
IMG_NI = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME_NI)
BINARY_NI = $(APP_NAME_NI)

.PHONY: build
build:
	./before-commit.sh ci

.PHONY: pull-licenses
pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

.PHONY: build-image
build-image: pull-licenses
	./docker-build-event-publish-service.sh $(APP_NAME_PUBLISH_KN) latest event-bus
	./docker-build-subscription-controller.sh $(APP_NAME_SCK) latest event-bus
	./docker-build-nats-init.sh $(APP_NAME_NI) latest event-bus

.PHONY: push-image
push-image:
	docker tag $(APP_NAME_PUBLISH_KN) $(IMG_PUBLISH_KN):$(TAG)
	docker push $(IMG_PUBLISH_KN):$(TAG)

	docker tag $(APP_NAME_SCK) $(IMG_SCK):$(TAG)
	docker push $(IMG_SCK):$(TAG)

	docker tag $(APP_NAME_NI) $(IMG_NI):$(TAG)
	docker push $(IMG_NI):$(TAG)

.PHONY: ci-pr
ci-pr: build build-image push-image

.PHONY: ci-master
ci-master: build build-image push-image

.PHONY: ci-release
ci-release: build build-image push-image

.PHONY: resolve
resolve:
	dep ensure -vendor-only -v

.PHONY: test
test:
	go list ./... | grep -v generated | xargs go test -v

.PHONY: clean
clean:
	rm -rf ./cmd/event-publish-service/$(BINARY_PUBLISH_KN) ./cmd/event-publish-service/docker
	rm -rf ./cmd/subscription-controller/$(BINARY_SCK) ./cmd/subscription-controller/docker
	rm -rf ./cmd/nats-streaming-init/$(BINARY_NI) ./cmd/nats-streaming-init/docker

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/event-bus"


#####################
#                   #
#  Code generation  #
#                   #
#####################

# see https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md#generate-code

PKG := github.com/kyma-project/kyma/components/event-bus

# e.g. "my-group/v1alpha1 my-group/v1alpha2"
API_GROUPS := eventing/v1alpha1 applicationconnector/v1alpha1
# generates e.g. "PKG/apis/sources/v1alpha1 PKG/apis/sources/v1alpha2"
api-import-paths := $(foreach group,$(API_GROUPS),$(PKG)/apis/$(group))

generators := deepcopy
generators_bin := $(foreach x,$(generators),bin/$(x)-gen)

.PHONY: codegen $(generators) $(generators_bin)

codegen: $(generators)

# http://blog.jgc.org/2007/06/escaping-comma-and-space-in-gnu-make.html
comma := ,
space :=
space +=

# doc: https://godoc.org/k8s.io/code-generator/cmd/deepcopy-gen
deepcopy: bin/deepcopy-gen
	@echo "+ Generating deepcopy funcs for $(API_GROUPS)"
	@bin/deepcopy-gen \
		--go-header-file hack/boilerplate.go.txt \
		--input-dirs $(subst $(space),$(comma),$(api-import-paths))

$(generators_bin): $(patched_trace)
	go build -o bin/$(notdir $@) ./vendor/k8s.io/code-generator/cmd/$(notdir $@)

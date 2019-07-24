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

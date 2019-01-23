TAG = $(DOCKER_TAG)

# event-bus-publish
APP_NAME_PUBLISH = event-bus-publish
IMG_PUBLISH = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME_PUBLISH)
BINARY_PUBLISH = $(APP_NAME_PUBLISH)

# event-bus-publish-knative
APP_NAME_PUBLISH_KN = event-bus-publish-knative
IMG_PUBLISH_KN = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME_PUBLISH_KN)
BINARY_PUBLISH_KN = $(APP_NAME_PUBLISH_KN)

# event-bus-push
APP_NAME_PUSH = event-bus-push
IMG_PUSH = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME_PUSH)
BINARY_PUSH = $(APP_NAME_PUSH)

# event-bus-subscription-controller-knative
APP_NAME_SCK = event-bus-subscription-controller-knative
IMG_SCK = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME_SCK)
BINARY_SCK = $(APP_NAME_SCK)

# event-bus-sv
APP_NAME_SV = event-bus-sub-validator
IMG_SV = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME_SV)
BINARY_SV = $(APP_NAME_SV)

.PHONY: build
build:
	# todo activate before-commit phase and remove the gometalinter
	#./before-commit.sh ci
	dep ensure -vendor-only -v
	go test ./...
	curl https://raw.githubusercontent.com/alecthomas/gometalinter/master/scripts/install.sh | sh -s v2.0.8
	./bin/gometalinter --skip=generated --vendor --deadline=2m --disable-all ./...

.PHONY: build-image
build-image:
	./docker-build-event-bus-publish.sh $(APP_NAME_PUBLISH) latest event-bus
	./docker-build-event-bus-publish-knative.sh $(APP_NAME_PUBLISH_KN) latest event-bus
	./docker-build-event-bus-push.sh $(APP_NAME_PUSH) latest event-bus
	./docker-build-event-bus-sck.sh $(APP_NAME_SCK) latest event-bus
	./docker-build-event-bus-sv.sh $(APP_NAME_SV) latest event-bus

.PHONY: push-image
push-image:
	docker tag $(APP_NAME_PUBLISH) $(IMG_PUBLISH):$(TAG)
	docker push $(IMG_PUBLISH):$(TAG)
	docker tag $(APP_NAME_PUBLISH_KN) $(IMG_PUBLISH_KN):$(TAG)
	docker push $(IMG_PUBLISH_KN):$(TAG)
	docker tag $(APP_NAME_PUSH) $(IMG_PUSH):$(TAG)
	docker push $(IMG_PUSH):$(TAG)
	docker tag $(APP_NAME_SCK) $(IMG_SCK):$(TAG)
	docker push $(IMG_SCK):$(TAG)
	docker tag $(APP_NAME_SV) $(IMG_SV):$(TAG)
	docker push $(IMG_SV):$(TAG)

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
	rm -rf ./cmd/event-bus-publish/$(BINARY_PUBLISH) ./cmd/event-bus-publish/docker
	rm -rf ./cmd/event-bus-publish-knative/$(BINARY_PUBLISH_KN) ./cmd/event-bus-publish-knative/docker
	rm -rf ./cmd/event-bus-push/$(BINARY_PUSH) ./cmd/event-bus-push/docker
	rm -rf ./cmd/event-bus-subscription-controller-knative/$(BINARY_SCK) ./cmd/event-bus-subscription-controller-knative/docker
	rm -rf ./cmd/event-bus-sv/$(BINARY_SV) ./cmd/event-bus-sv/docker

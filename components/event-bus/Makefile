TAG = $(DOCKER_TAG)

# event-bus-publish
APP_NAME_PUBLISH = event-bus-publish
IMG_PUBLISH = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME_PUBLISH)
BINARY_PUBLISH = $(APP_NAME_PUBLISH)

# event-bus-push
APP_NAME_PUSH = event-bus-push
IMG_PUSH = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME_PUSH)
BINARY_PUSH = $(APP_NAME_PUSH)

# event-bus-sv
APP_NAME_SV = event-bus-sub-validator
IMG_SV = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME_SV)
BINARY_SV = $(APP_NAME_SV)

.PHONY: build
build:
	# todo activate before-commit phase and remove the gometalinter
	#./before-commit.sh ci
	dep ensure -vendor-only -v
	curl https://raw.githubusercontent.com/alecthomas/gometalinter/master/scripts/install.sh | sh -s v2.0.8
	./bin/gometalinter --skip=generated --vendor --deadline=2m --disable-all ./...

.PHONY: build-image
build-image:
	./docker-build-event-bus-publish.sh $(APP_NAME_PUBLISH) latest event-bus
	./docker-build-event-bus-push.sh $(APP_NAME_PUSH) latest event-bus
	./docker-build-event-bus-sv.sh $(APP_NAME_SV) latest event-bus

.PHONY: push-image
push-image:
	docker tag $(APP_NAME_PUBLISH) $(IMG_PUBLISH):$(TAG)
	docker push $(IMG_PUBLISH):$(TAG)
	docker tag $(APP_NAME_PUSH) $(IMG_PUSH):$(TAG)
	docker push $(IMG_PUSH):$(TAG)
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
	rm -f ./cmd/event-bus-publish/$(BINARY_PUBLISH)
	rm -f ./cmd/event-bus-push/$(BINARY_PUSH)
	rm -f ./cmd/event-bus-sv/$(BINARY_SV)

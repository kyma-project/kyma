SUBSCRIBER_APP_NAME = event-bus-e2e-subscriber
TESTER_APP_NAME = event-bus-e2e-tester
SUBSCRIBER_IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(SUBSCRIBER_APP_NAME)
TESTER_IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(TESTER_APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: resolve
resolve:
	dep ensure -vendor-only -v

.PHONY: build
build:
	./before-commit.sh ci

.PHONY: build-image
build-image:
	./docker-build-subscriber.sh ${SUBSCRIBER_APP_NAME} latest ${SUBSCRIBER_APP_NAME}
	./docker-build-tester.sh ${TESTER_APP_NAME} latest ${TESTER_APP_NAME}

.PHONY: push-image
push-image:
	docker tag $(SUBSCRIBER_APP_NAME) $(SUBSCRIBER_IMG):$(TAG)
	docker tag $(TESTER_APP_NAME) $(TESTER_IMG):$(TAG)
	docker push $(SUBSCRIBER_IMG):$(TAG)
	docker push $(TESTER_IMG):$(TAG)

.PHONY: ci-pr
ci-pr: build build-image push-image

.PHONY: ci-master
ci-master: build build-image push-image

.PHONY: ci-release
ci-release: build build-image push-image

.PHONY: clean
clean:
	rm -rf  ./e2e-subscriber/bin/
	rm -rf ./e2e-subscriber/docker/
	rm -rf  ./e2e-tester/bin/
	rm -rf ./e2e-tester/docker/

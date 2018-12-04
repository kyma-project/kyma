APP_NAME = stability-checker
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: build
build:
	./before-commit.sh ci

.PHONY: build-image
build-image:
	cp $(APP_NAME) deploy/stability-checker/$(APP_NAME)
	cp logs-printer deploy/stability-checker/logs-printer
	docker build -t $(APP_NAME):latest deploy/stability-checker

.PHONY: push-image
push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

.PHONY: ci-pr
ci-pr: build build-image push-image

.PHONY: ci-master
ci-master: build build-image push-image

.PHONY: ci-release
ci-release: build build-image push-image

.PHONY: clean
clean:
	rm -f $(APP_NAME)
	rm -f logs-printer


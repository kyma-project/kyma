APP_NAME = application-gateway-tests
TEST_EXECUTOR_APP_NAME = application-gateway-tests-executor

IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TEST_EXECUTOR_IMAGE = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(TEST_EXECUTOR_APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: build
build:
	./before-commit.sh

.PHONY: pull-licenses
pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

.PHONY: build-image
build-image: pull-licenses
	docker build -t $(APP_NAME):latest . -f ./test/helmtest/Dockerfile --build-arg TEST_EXECUTOR_IMAGE=$(TEST_EXECUTOR_IMAGE):$(TAG)
	docker build -t $(TEST_EXECUTOR_APP_NAME):latest . -f ./test/executor/Dockerfile

.PHONY: push-image
push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker tag $(TEST_EXECUTOR_APP_NAME) $(TEST_EXECUTOR_IMAGE):$(TAG)
	docker push $(IMG):$(TAG)
	docker push $(TEST_EXECUTOR_IMAGE):$(TAG)

.PHONY: ci-pr
ci-pr: build build-image push-image

.PHONY: ci-master
ci-master: build build-image push-image

.PHONY: ci-release
ci-release: build build-image push-image

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"
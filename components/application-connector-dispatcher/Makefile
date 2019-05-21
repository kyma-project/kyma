VALIDATION_PROXY_APP_NAME = application-connector-validation-proxy
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(VALIDATION_PROXY_APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: build
build:
	./before-commit.sh ci

.PHONY: build-image
build-image:
	docker build -t $(VALIDATION_PROXY_APP_NAME):latest . -f Dockerfile.validation-proxy

.PHONY: push-image
push-image:
	docker tag $(VALIDATION_PROXY_APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

.PHONY: ci-pr
ci-pr: build build-image push-image

.PHONY: ci-master
ci-master: build build-image push-image

.PHONY: ci-release
ci-release: build build-image push-image

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

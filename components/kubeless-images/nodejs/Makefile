
APP_NAME = kubeless-images
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

build-image:
	docker build -t $(APP_NAME)-node6:latest --build-arg NODEIMAGE='kubeless/nodejs@sha256:4d8f5d6ad51676e8faffc0da6fabe8e12cf8cf34ae0ce8e94ee4e510f9876d2e' .
	docker build -t $(APP_NAME)-node8:latest --build-arg NODEIMAGE='kubeless/nodejs@sha256:5f1e999a1021dfb3d117106d80519a82110bd26a579f067f1ff7127025c90be5' .

push-image:
	docker tag $(APP_NAME)-node6:latest $(IMG)-node6:$(TAG)
	docker push $(IMG)-node6:$(TAG)

	docker tag $(APP_NAME)-node8:latest $(IMG)-node8:$(TAG)
	docker push $(IMG)-node8:$(TAG)



.PHONY: ci-pr
ci-pr: build-image push-image

.PHONY: ci-master
ci-master: build-image push-image

.PHONY: ci-release
ci-release: build-image push-image
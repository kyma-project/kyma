
APP_NAME = kubeless-images
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

build-image:
	docker build -t $(APP_NAME)-node6:latest  -f Dockerfile.6 .
	docker build -t $(APP_NAME)-node8:latest  -f Dockerfile.8 .

push-image:
	docker tag $(APP_NAME)-node6:latest $(IMG)-node6:latest:$(TAG)
	docker push $(IMG):$(TAG)



.PHONY: ci-pr
ci-pr: resolve validate build build-image push-image

.PHONY: ci-master
ci-master: resolve validate build build-image push-image

.PHONY: ci-release
ci-release: resolve validate build build-image push-image
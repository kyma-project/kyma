# This Makefile does not use the generic file located at $(SCRIPTS_DIR)/generic-make-go.mk. 
# It just provides the `release`-goal 
#
APP_NAME := kubeless-images
IMG := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG := $(DOCKER_TAG)

build-image: build-image-node6 build-image-node8 ;
push-image: push-image-node6 push-image-node8 ;

build-image-node6:
	docker build -t $(APP_NAME)-node6:latest --build-arg NODEIMAGE='kubeless/nodejs@sha256:4d8f5d6ad51676e8faffc0da6fabe8e12cf8cf34ae0ce8e94ee4e510f9876d2e' .

build-image-node8:
	docker build -t $(APP_NAME)-node8:latest --build-arg NODEIMAGE='kubeless/nodejs@sha256:5f1e999a1021dfb3d117106d80519a82110bd26a579f067f1ff7127025c90be5' .

push-image-node6:
	docker tag $(APP_NAME)-node6:latest $(IMG)-node6:$(TAG)
	docker push $(IMG)-node6:$(TAG)

push-image-node8:
	docker tag $(APP_NAME)-node8:latest $(IMG)-node8:$(TAG)
	docker push $(IMG)-node8:$(TAG)

release: build-image push-image

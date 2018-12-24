APP_NAME = kyma-installer
DOCKER_PATH=$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: ci-release
ci-release:
	cd ./../../; \
	docker build -t $(IMG):$(TAG) \
		-f ./tools/kyma-installer/kyma.Dockerfile  \
		--build-arg INSTALLER_DIR=$(DOCKER_PATH) \
		--build-arg INSTALLER_VERSION=$(TAG) .
	docker push $(IMG):$(TAG)
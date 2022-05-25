# Environment variables expected by this makefile
# DOCKER_TAG - dynamic tag for image
# DOCKER_PUSH_REPOSITORY
# DOCKER_PUSH_DIRECTORY

IMG_NAME = kyma-fi-test
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)

.PHONY: ci-pr
ci-pr: build-image push-image

.PHONY: build-image
build-image:
	$(info ___ Building image $(IMG) ___)
	docker build . -t $(IMG):$(DOCKER_TAG)

.PHONY: push-image
push-image:
	$(info ___ Pushing image $(IMG) with all tags ___)
	docker push -a $(IMG)

.PHONY: release
release: build-image push-image


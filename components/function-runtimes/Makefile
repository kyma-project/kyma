ROOT :=  $(shell pwd)

release: build-images push-images

.PHONY: build-images
	build-function-runtime-nodejs16-image \
	build-function-runtime-python39-image

.PHONY: push-images
	push-function-runtime-nodejs16-image \
	push-function-runtime-python39-image


## nodejs16
FUNCTION_RUNTIME_IMAGE_NODEJS16 = function-runtime-nodejs16
FUNCTION_RUNTIME_NODEJS16_IMG_NAME:=$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(FUNCTION_RUNTIME_IMAGE_NODEJS16)

.PHONY: build-function-runtime-nodejs16-image push-function-runtime-nodejs16-image
build-function-runtime-nodejs16-image:
	docker build -t $(FUNCTION_RUNTIME_IMAGE_NODEJS16) $(ROOT)/nodejs16/

push-function-runtime-nodejs16-image:
	docker tag $(FUNCTION_RUNTIME_IMAGE_NODEJS16) $(FUNCTION_RUNTIME_NODEJS16_IMG_NAME):$(DOCKER_TAG)
	docker push $(FUNCTION_RUNTIME_NODEJS16_IMG_NAME):$(DOCKER_TAG)
ifeq ($(JOB_TYPE), postsubmit)
	@echo "Sign image with Cosign"
	cosign version
	cosign sign -key ${KMS_KEY_URL} $(FUNCTION_RUNTIME_NODEJS16_IMG_NAME):$(DOCKER_TAG)
else
	@echo "Image signing skipped"
endif

## python39
FUNCTION_RUNTIME_IMAGE_PYTHON39 = function-runtime-python39
FUNCTION_RUNTIME_PYTHON39_IMG_NAME:=$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(FUNCTION_RUNTIME_IMAGE_PYTHON39)

.PHONY: build-function-runtime-python39-image push-function-runtime-python39-image
build-function-runtime-python39-image:
	docker build -t $(FUNCTION_RUNTIME_IMAGE_PYTHON39) $(ROOT)/python39/

push-function-runtime-python39-image:
	docker tag $(FUNCTION_RUNTIME_IMAGE_PYTHON39) $(FUNCTION_RUNTIME_PYTHON39_IMG_NAME):$(DOCKER_TAG)
	docker push $(FUNCTION_RUNTIME_PYTHON39_IMG_NAME):$(DOCKER_TAG)
ifeq ($(JOB_TYPE), postsubmit)
	@echo "Sign image with Cosign"
	cosign version
	cosign sign -key ${KMS_KEY_URL} $(FUNCTION_RUNTIME_PYTHON39_IMG_NAME):$(DOCKER_TAG)
else
	@echo "Image signing skipped"
endif

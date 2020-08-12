ROOT :=  $(shell pwd)

release: build-images push-images

.PHONY: build-images
build-images: build-function-runtime-nodejs12-image \
	build-function-runtime-nodejs10-image \
	build-function-runtime-python37-image

.PHONY: push-images
push-images: push-function-runtime-nodejs12-image \
	push-function-runtime-nodejs10-image \
	push-function-runtime-python37-image

## nodejs12
FUNCTION_RUNTIME_IMAGE_NODEJS12 = function-runtime-nodejs12
FUNCTION_RUNTIME_NODEJS12_IMG_NAME:=$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(FUNCTION_RUNTIME_IMAGE_NODEJS12)

.PHONY: build-function-runtime-nodejs12-image push-function-runtime-nodejs12-image
build-function-runtime-nodejs12-image:
	docker build -t $(FUNCTION_RUNTIME_IMAGE_NODEJS12) $(ROOT)/nodejs12/

push-function-runtime-nodejs12-image:
	docker tag $(FUNCTION_RUNTIME_IMAGE_NODEJS12) $(FUNCTION_RUNTIME_NODEJS12_IMG_NAME):$(DOCKER_TAG)
	docker push $(FUNCTION_RUNTIME_NODEJS12_IMG_NAME):$(DOCKER_TAG)

## nodejs10
FUNCTION_RUNTIME_IMAGE_NODEJS10 = function-runtime-nodejs10
FUNCTION_RUNTIME_NODEJS10_IMG_NAME:=$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(FUNCTION_RUNTIME_IMAGE_NODEJS10)

.PHONY: build-function-runtime-nodejs10-image push-function-runtime-nodejs10-image
build-function-runtime-nodejs10-image:
	docker build -t $(FUNCTION_RUNTIME_IMAGE_NODEJS10) $(ROOT)/nodejs10/

push-function-runtime-nodejs10-image:
	docker tag $(FUNCTION_RUNTIME_IMAGE_NODEJS10) $(FUNCTION_RUNTIME_NODEJS10_IMG_NAME):$(DOCKER_TAG)
	docker push $(FUNCTION_RUNTIME_NODEJS10_IMG_NAME):$(DOCKER_TAG)

## python37
FUNCTION_RUNTIME_IMAGE_PYTHON37 = function-runtime-python37
FUNCTION_RUNTIME_PYTHON37_IMG_NAME:=$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(FUNCTION_RUNTIME_IMAGE_PYTHON37)

.PHONY: build-function-runtime-python37-image push-function-runtime-python37-image
build-function-runtime-python37-image:
	docker build -t $(FUNCTION_RUNTIME_IMAGE_PYTHON37) $(ROOT)/python37/

push-function-runtime-python37-image:
	docker tag $(FUNCTION_RUNTIME_IMAGE_PYTHON37) $(FUNCTION_RUNTIME_PYTHON37_IMG_NAME):$(DOCKER_TAG)
	docker push $(FUNCTION_RUNTIME_PYTHON37_IMG_NAME):$(DOCKER_TAG)
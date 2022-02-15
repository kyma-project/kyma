APP_NAME = istio-installer
APP_PATH = components/$(APP_NAME)
IMG_NAME = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: build-image
build-image:
	docker build -t $(APP_NAME):latest .

.PHONY: push-image
push-image:
	docker tag $(APP_NAME) $(IMG_NAME):$(TAG)
	docker push $(IMG_NAME):$(TAG)
ifeq ($(JOB_TYPE), postsubmit)
	@echo "Sign image with Cosign"
	cosign version
	cosign sign -key ${KMS_KEY_URL} $(IMG_NAME):$(TAG)
else
	@echo "Image signing skipped"
endif

release: build-image push-image

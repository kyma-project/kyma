APP_NAME = test-namespace-controller
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: prepare
prepare:
	./prepare.sh

.PHONY: pull-licenses
pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

.PHONY: build-image
build-image: pull-licenses
	docker build -t $(APP_NAME):latest .

.PHONY: push-image
push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

.PHONY: ci-pr
ci-pr: prepare build-image push-image

.PHONY: ci-master
ci-master: prepare build-image push-image

.PHONY: ci-release
ci-release: prepare build-image push-image

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

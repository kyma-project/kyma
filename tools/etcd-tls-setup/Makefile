APP_NAME = etcd-tls-setup
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)
binary=$(APP_NAME)

.PHONY: build-image
build-image:
	docker build -t $(APP_NAME):latest .

.PHONY: push-image
push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

.PHONY: ci-pr
ci-pr: build-image push-image

.PHONY: ci-master
ci-master: build-image push-image

.PHONY: ci-release
ci-release: build-image push-image

.PHONY: clean
clean:
	rm -f $(binary)


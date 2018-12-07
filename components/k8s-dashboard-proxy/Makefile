APP_NAME = k8s-dashboard-proxy
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)
binary = reverseproxy

.PHONY: compile
compile:
	mkdir -p ./bin
	go build ./cmd/reverseproxy
	mv ./reverseproxy ./bin

.PHONY: build
build:
	./before-commit.sh ci

.PHONY: build-image
build-image:
	docker build -t $(APP_NAME):latest .

.PHONY: push-image
push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

.PHONY: ci-pr
ci-pr: build build-image push-image

.PHONY: ci-master
ci-master: build build-image push-image

.PHONY: ci-release
ci-release: build build-image push-image

.PHONY: clean
clean:
	rm -rf docker/
	rm -f $(binary)

REGISTRY = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)
TAG = $(DOCKER_TAG)

.PHONY: ci-pr ci-master ci-release resolve build-and-test build-image push-image pull-licenses

ci-master: push-image
ci-pr: ci-master
ci-release: ci-master

pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

build-and-test:
	./before-commit.sh ci

# TODO: Iterate over dockerfiles in deploy/*
build-image: build-and-test pull-licenses
	docker build -t cms-asyncapi-service -f deploy/asyncapi/Dockerfile .

push-image: build-image
	docker tag cms-asyncapi-service $(REGISTRY)/cms-asyncapi-service:$(TAG)
	docker push $(REGISTRY)/cms-asyncapi-service:$(TAG)

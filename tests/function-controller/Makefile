.PHONY: build test

build:
	@echo '+ Performing pre-commit checks'
	@./before-commit.sh

test:
	@echo '+ Running e2e tests'
ifdef TEST_ARGS
	@go test ./ -args $(TEST_ARGS)
else
	@go test ./
endif

vendor: Gopkg.lock
	@echo '+ Pulling vendored dependencies'
	@dep ensure -v --vendor-only

# CI targets

APP_NAME := function-controller-e2e-tests
IMG := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG := $(DOCKER_TAG)

.PHONY: build-image push-image pull-licenses ci-pr ci-master ci-release

build-image: pull-licenses
	docker build -t $(APP_NAME):latest .

push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

ci-pr: build build-image push-image

ci-master: build build-image push-image

ci-release: build build-image push-image

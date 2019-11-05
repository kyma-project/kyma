.PHONY: build test

build:
	@echo '+ Performing pre-commit checks'
	@./before-commit.sh

# The 0 timeout prevents the test from panicking after 10m, leaving stale resources behind.
test: vendor/k8s.io/kubernetes/test/e2e/framework/create.go
	@echo '+ Running e2e tests'
ifdef TEST_ARGS
	@go test -timeout=0 -v ./ -args $(TEST_ARGS)
else
	@go test -timeout=0 -v ./
endif

# This patch adds support for few extra types to the 'CreateFromManifests' util function.
vendor/k8s.io/kubernetes/test/e2e/framework/create.go: hack/k8s_framework_create.patch vendor
	@echo '+ Patching vendored kubernetes/test/e2e/framework'
	@patch -p0 < hack/k8s_framework_create.patch

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

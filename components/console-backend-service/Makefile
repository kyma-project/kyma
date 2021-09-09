ENTRYPOINT := ./main.go
TAG := $(DOCKER_TAG)

APP_NAME = console-backend-service
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
IMG_NAME := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)

# VERIFY_IGNORE is a grep pattern to exclude files and directories from verification
VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg
# FILES_TO_CHECK is a command used to determine which files should be verified
FILES_TO_CHECK = find . -type f -name "*.go" | grep -v "$(VERIFY_IGNORE)"
# DIRS_TO_CHECK is a command used to determine which directories should be verified
DIRS_TO_CHECK = go list ./... | grep -v "$(VERIFY_IGNORE)"


release: build-image push-image

build-image: pull-licenses
	docker build -t $(APP_NAME) .
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

pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

format: imports fmt

verify: test check-imports check-fmt

test:
	mkdir -p /tmp/artifacts
	go test -coverprofile=/tmp/artifacts/cover.out ./...
	@echo -n "Total coverage: "
	@go tool cover -func=/tmp/artifacts/cover.out | grep total | awk '{print $$3}'

check-imports:
	@if [ -n "$$(goimports -l $$($(FILES_TO_CHECK)))" ]; then \
		echo "✗ some files contain not propery formatted imports. To repair run make imports-local"; \
		goimports -l $$($(FILES_TO_CHECK)); \
		exit 1; \
	fi;

check-fmt:
	@if [ -n "$$(gofmt -l $$($(FILES_TO_CHECK)))" ]; then \
		gofmt -l $$($(FILES_TO_CHECK)); \
		echo "✗ some files contain not propery formatted imports. To repair run make imports-local"; \
		exit 1; \
	fi;

imports:
	goimports -w -l $$($(FILES_TO_CHECK))

fmt:
	go fmt $$($(DIRS_TO_CHECK))

build:
	env CGO_ENABLED=0 go build -o $(APP_NAME) ./$(ENTRYPOINT)

run: export APP_KUBECONFIG_PATH=/Users/${USER}/.kube/config
run: export APP_VERBOSE=true
run: export APP_RAFTER_ADDRESS=https://storage.${KYMA_DOMAIN}
run: export APP_RAFTER_VERIFY_SSL=false
run: export APP_APPLICATION_GATEWAY_INTEGRATION_NAMESPACE=kyma-integration
run: export APP_APPLICATION_CONNECTOR_URL=http://dummy.url
run: export APP_OIDC_ISSUER_URL=https://dex.${KYMA_DOMAIN}
run: export APP_OIDC_CLIENT_ID=kyma-client

run:
	./console-backend-service

watch:
	CompileDaemon -build="make build" -command="make run" -graceful-kill

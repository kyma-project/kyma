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

.PHONY: release
release: build-image push-image

build-image: pull-licenses
	docker build -t $(IMG_NAME) .
push-image:
	docker tag $(IMG_NAME) $(IMG_NAME):$(TAG)
	docker push $(IMG_NAME):$(TAG)

pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

verify: test check-imports check-fmt

test-local:
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

build:
	env CGO_ENABLED=0 go build -o $(APP_NAME) ./$(ENTRYPOINT)
	rm $(APP_NAME)

watch: export APP_KUBECONFIG_PATH=/Users/${USER}/.kube/config
watch: export APP_VERBOSE=true
watch: export APP_RAFTER_ADDRESS=https://storage.${KYMA_DOMAIN}
watch: export APP_RAFTER_VERIFY_SSL=false
watch: export APP_APPLICATION_GATEWAY_INTEGRATION_NAMESPACE=kyma-integration
watch: export APP_APPLICATION_CONNECTOR_URL=http://dummy.url
watch: export APP_OIDC_ISSUER_URL=https://dex.${KYMA_DOMAIN}
watch: export APP_OIDC_CLIENT_ID=kyma-client

watch:
	CompileDaemon -build="make build-local" -command="go run main.go"
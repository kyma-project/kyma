TAG := $(DOCKER_TAG)
APP_NAME = common
APP_PATH = common
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
IMG_NAME := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)

# VERIFY_IGNORE is a grep pattern to exclude files and directories from verification
VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg
# FILES_TO_CHECK is a command used to determine which files should be verified
FILES_TO_CHECK = find . -type f -name "*.go" | grep -v "$(VERIFY_IGNORE)"
# DIRS_TO_CHECK is a command used to determine which directories should be verified
DIRS_TO_CHECK = go list ./... | grep -v "$(VERIFY_IGNORE)"

release:
	@echo "-"

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
	env CGO_ENABLED=0 go build -o $(APP_NAME) ./

run:
	./common

watch:
	CompileDaemon -build="make build" -command="make run" -graceful-kill
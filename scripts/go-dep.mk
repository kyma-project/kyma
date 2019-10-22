# Default configuration
MAIN_PATH := ./main.go
IMG_NAME := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG := $(DOCKER_TAG)
# BASE_PKG is a root packge of the component
BASE_PKG := github.com/kyma-project/kyma
# IMG_GOPATH is a path to go path in the container
IMG_GOPATH := /workspace/go
# IMG_GOCACHE is a path to go cache in the container
IMG_GOCACHE := /root/.cache/go-build
# VERIFY_IGNORE is a grep pattern to exclude files and directories from verification
VERIFY_IGNORE := /vendor\|/automock

# Other variables
# LOCAL_DIR in a local path to scripts folder
LOCAL_DIR = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
# COMPONENT_DIR is a local path to commponent
COMPONENT_DIR = $(shell pwd)
# WORKSPACE_LOCAL_DIR is a path to the scripts folder in the container
WORKSPACE_LOCAL_DIR = $(IMG_GOPATH)/src/$(BASE_PKG)/scripts
# WORKSPACE_COMPONENT_DIR is a path to commponent in hte container
WORKSPACE_COMPONENT_DIR = $(IMG_GOPATH)/src/$(BASE_PKG)/$(APP_PATH)
# FILES_TO_CHECK is a command used to determine which files should be verified
FILES_TO_CHECK = find . -type f -name "*.go" | grep -v "$(VERIFY_IGNORE)"
# DIRS_TO_CHECK is a command used to determine which directories should be verified
DIRS_TO_CHECK = go list ./... | grep -v "$(VERIFY_IGNORE)"
# DIRS_TO_IGNORE is a command used to determine which directories should not be verified
DIRS_TO_IGNORE = go list ./... | grep "$(VERIFY_IGNORE)"

# Base docker configuration
DOCKER_CREATE_OPTS := -v $(LOCAL_DIR):$(WORKSPACE_LOCAL_DIR):delegated --rm -w $(WORKSPACE_COMPONENT_DIR) $(BUILDPACK)

# Check if go is available
ifneq (,$(shell go version 2>/dev/null))
DOCKER_CREATE_OPTS := -v $(shell go env GOCACHE):$(IMG_GOCACHE):delegated -v $(shell go env GOPATH)/pkg/dep:$(IMG_GOPATH)/pkg/dep:delegated $(DOCKER_CREATE_OPTS)
endif

# Check if running with TTY
ifeq (1, $(shell [ -t 0 ] && echo 1))
DOCKER_INTERACTIVE := -i
DOCKER_CREATE_OPTS := -t $(DOCKER_CREATE_OPTS)
else
DOCKER_INTERACTIVE_START := --attach 
endif

# Buildpack directives
define buildpack-mount
.PHONY: $(1)-local $(1)
$(1):
	@echo make $(1)
	@docker run $(DOCKER_INTERACTIVE) \
		-v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
		$(DOCKER_CREATE_OPTS) make $(1)-local
endef

define buildpack-cp-ro
.PHONY: $(1)-local $(1)
$(1):
	@echo make $(1)
	$$(eval container = $$(shell docker create $(DOCKER_CREATE_OPTS) make $(1)-local))
	@docker cp $(COMPONENT_DIR)/. $$(container):$(WORKSPACE_COMPONENT_DIR)/
	@docker start $(DOCKER_INTERACTIVE_START) $(DOCKER_INTERACTIVE) $$(container)
endef

.PHONY: verify format release

# You may add additional targets/commands to be run on format and verify. Declare the target again in your makefile,
# using two double colons. For example to run errcheck on verify add this to your makefile:
#
#   verify:: errcheck
#
verify:: test check-imports check-fmt
format:: imports fmt

release: resolve dep-status verify build-image push-image

.PHONY: build-image push-image
build-image: pull-licenses
	docker build -t $(IMG_NAME) .
push-image:
	docker tag $(IMG_NAME) $(IMG_NAME):$(TAG)
	docker push $(IMG_NAME):$(TAG)
docker-create-opts:
	@echo $(DOCKER_CREATE_OPTS)

# Targets mounting sources to buildpack
MOUNT_TARGETS = build resolve ensure dep-status check-imports imports check-fmt fmt errcheck vet generate pull-licenses
$(foreach t,$(MOUNT_TARGETS),$(eval $(call buildpack-mount,$(t))))

build-local:
	env CGO_ENABLED=0 go build -o $(APP_NAME) $(MAIN_PATH)
	rm $(APP_NAME)

resolve-local:
	dep ensure -vendor-only -v

ensure-local:
	dep ensure -v

dep-status-local:
	dep status -v

check-imports-local:
	exit $(shell goimports -l $$($(FILES_TO_CHECK)) | wc -l | xargs)

imports-local:
	goimports -w -l $$($(FILES_TO_CHECK))

check-fmt-local:
	exit $(shell gofmt -l $$($(FILES_TO_CHECK)) | wc -l | xargs)

fmt-local:
	go fmt $$($(DIRS_TO_CHECK))

errcheck-local:
	errcheck -blank -asserts -ignorepkg '$$($(DIRS_TO_IGNORE) | tr '\n' ',')' -ignoregenerated ./...

vet-local:
	go vet $$($(DIRS_TO_CHECK))

generate-local:
	go generate ./...

pull-licenses-local:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

# Targets copying sources to buildpack
COPY_TARGETS = test
$(foreach t,$(COPY_TARGETS),$(eval $(call buildpack-cp-ro,$(t))))

test-local:
	go test ./...

.PHONY: list
list:
	@$(MAKE) -pRrq -f $(COMPONENT_DIR)/Makefile : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'

.PHONY: exec
exec:
	@docker run $(DOCKER_INTERACTIVE) \
    		-v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
    		$(DOCKER_CREATE_OPTS) bash
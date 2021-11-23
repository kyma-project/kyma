# Default configuration
ENTRYPOINT := ./main.go
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

# IGNORE_LINTING_ISSUES enables or disables ignoring linting. 
# enable failing builds by specifying the following in the components makefile
# override IGNORE_LINTING_ISSUES =
IGNORE_LINTING_ISSUES := -

# Other variables
# LOCAL_DIR in a local path to scripts folder
LOCAL_DIR = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
# COMPONENT_DIR is a local path to commponent
COMPONENT_DIR = $(shell pwd)
# WORKSPACE_LOCAL_DIR is a path to the scripts folder in the container
WORKSPACE_LOCAL_DIR = $(IMG_GOPATH)/src/$(BASE_PKG)/common/makefiles
# WORKSPACE_COMPONENT_DIR is a path to commponent in hte container
WORKSPACE_COMPONENT_DIR = $(IMG_GOPATH)/src/$(BASE_PKG)/$(APP_PATH)
# FILES_TO_CHECK is a command used to determine which files should be verified
FILES_TO_CHECK = find . -type f -name "*.go" | grep -v "$(VERIFY_IGNORE)"
# DIRS_TO_CHECK is a command used to determine which directories should be verified
DIRS_TO_CHECK = go list ./... | grep -v "$(VERIFY_IGNORE)"
# DIRS_TO_IGNORE is a command used to determine which directories should not be verified
DIRS_TO_IGNORE = go list ./... | grep "$(VERIFY_IGNORE)"

ifndef ARTIFACTS
ARTIFACTS:=/tmp/artifacts
endif

# Base docker configuration
DOCKER_CREATE_OPTS := -v $(LOCAL_DIR):$(WORKSPACE_LOCAL_DIR):delegated -v $(ARTIFACTS):/tmp/artifacts --rm -w $(WORKSPACE_COMPONENT_DIR) $(BUILDPACK)

# Check if go is available
ifneq (,$(shell go version 2>/dev/null))
DOCKER_CREATE_OPTS := -v $(shell go env GOCACHE):$(IMG_GOCACHE):delegated -v $(shell go env GOPATH)/pkg/dep:$(IMG_GOPATH)/pkg/dep:delegated $(DOCKER_CREATE_OPTS)
endif

.DEFAULT_GOAL := verify

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

##@ Common Targets (not all might be useful for your project)
##@ Most targets with `-local` suffix can be executed without `-local` (e.g. `resolve-local` => `resolve`). If those targets are called without the `-local` suffix, then they will be executed inside a docker container with a preconfigured build environment.


.PHONY: verify format release check-gqlgen

# You may add additional targets/commands to be run on format and verify. Declare the target again in your makefile,
# using two double colons. For example to run errcheck on verify add this to your makefile:
#
#   verify:: errcheck
#
verify:: test check-imports check-fmt lint ## Check formatting and run linters
format:: imports fmt


##@ Common Release
release: ## Release the component
	$(MAKE) release-dep

gomod-release:gomod-component-check build-image push-image ## Build a go mod based project. Build the image and push it. Requires DOCKER_PUSH_REPOSITORY DOCKER_PUSH_DIRECTORY

gomod-release-local:gomod-component-check-local build-image push-image


#Old Target for dep projects
release-dep: resolve dep-status verify build-image push-image ## Release dep based project


##@ Common Docker
.PHONY: build-image push-image
build-image: pull-licenses-local ## Build the docker image
	docker build -t $(IMG_NAME) .
push-image: post-pr-tag-image ## Build and push the docker image. Needs DOCKER_PUSH_REPOSITORY DOCKER_PUSH_DIRECTORY 
	docker tag $(IMG_NAME) $(IMG_NAME):$(TAG)
	docker push $(IMG_NAME):$(TAG)
ifeq ($(JOB_TYPE), postsubmit)
	@echo "Sign image with Cosign"
	cosign version
	cosign sign -key ${KMS_KEY_URL} $(IMG_NAME):$(TAG)
else
	@echo "Image signing skipped"
endif
docker-create-opts:
	@echo $(DOCKER_CREATE_OPTS)
.PHONY: post-pr-tag-image
post-pr-tag-image:
ifdef DOCKER_POST_PR_TAG
	docker tag $(IMG_NAME) $(IMG_NAME):$(DOCKER_POST_PR_TAG)
endif

# Targets mounting sources to buildpack
MOUNT_TARGETS = build resolve ensure dep-status check-imports imports check-fmt fmt errcheck vet generate gqlgen
$(foreach t,$(MOUNT_TARGETS),$(eval $(call buildpack-mount,$(t))))

build-local: ## Build the binary
	env CGO_ENABLED=0 go build -o $(APP_NAME) ./$(ENTRYPOINT)
	rm $(APP_NAME)


##@ Common dep based builds
resolve-local: ## download `dep` dependencies
	dep ensure -vendor-only -v

ensure-local: ## Run `dep ensure`
	dep ensure -v

dep-status-local: ## Check `dep` status
	dep status -v

##@ Common go mod based builds
gomod-deps-local:: gomod-vendor-local gomod-verify-local ## Download gomod dependencies
$(eval $(call buildpack-mount,gomod-deps))

gomod-check-local:: test-local check-imports-local check-fmt-local lint ## Run tests and code checkers
$(eval $(call buildpack-cp-ro,gomod-check))

gomod-component-check-local:: gomod-deps-local gomod-check-local ## Download dependencies, run tests and checks
$(eval $(call buildpack-mount,gomod-component-check))


gomod-vendor-local:
	GO111MODULE=on go mod vendor

gomod-verify-local:
	GO111MODULE=on go mod verify

gomod-tidy-local:
	GO111MODULE=on go mod tidy

##@ Common Source Code tools
check-imports-local: ## Check import sections
	@if [ -n "$$(goimports -l $$($(FILES_TO_CHECK)))" ]; then \
		echo "✗ some files contain not propery formatted imports. To repair run make imports-local"; \
		goimports -l $$($(FILES_TO_CHECK)); \
		exit 1; \
	fi;

imports-local: ## Optimize imports
	goimports -w -l $$($(FILES_TO_CHECK))

check-fmt-local: ## Check source files for formatting issues
	@if [ -n "$$(gofmt -l $$($(FILES_TO_CHECK)))" ]; then \
		gofmt -l $$($(FILES_TO_CHECK)); \
		echo "✗ some files contain not propery formatted imports. To repair run make imports-local"; \
		exit 1; \
	fi;

fmt-local: ## Reformat files using `go fmt`
	go fmt $$($(DIRS_TO_CHECK))

errcheck-local:
	errcheck -blank -asserts -ignorepkg '$$($(DIRS_TO_CHECK) | tr '\n' ',')' -ignoregenerated ./...

vet-local:
	go vet $$($(DIRS_TO_CHECK))

# Lint goal is by default optional. To force failing this target reconfigure `IGNORE_LINTING_ISSUES` in your components Makefile
# Forcing build failing:
# override IGNORE_LINTING_ISSUES = 
lint: ## Run various linters
	$(IGNORE_LINTING_ISSUES)SKIP_VERIFY="true" ../../hack/verify-lint.sh $(COMPONENT_DIR) 

generate-local: ## Run code generation
	go generate ./...

gqlgen-local:
	./gqlgen.sh

check-gqlgen:
	@echo make gqlgen-check
	@if [ -n "$$(git status -s pkg/graphql)" ]; then \
		echo -e "${RED}✗ gqlgen.sh modified some files, schema and code are out-of-sync${NC}"; \
		git status -s pkg/graphql; \
		exit 1; \
	fi;

pull-licenses: pull-licenses-local

pull-licenses-local: ## Download licenses to store them in the docker image
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

# Targets copying sources to buildpack
COPY_TARGETS = test
$(foreach t,$(COPY_TARGETS),$(eval $(call buildpack-cp-ro,$(t))))

test-local: ## Run tests
	mkdir -p /tmp/artifacts
	go test -coverprofile=/tmp/artifacts/cover.out ./...
	@echo -n "Total coverage: "
	@go tool cover -func=/tmp/artifacts/cover.out | grep total | awk '{print $$3}'


##@ Other

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n" } /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST) <(printf "\
		")

.PHONY: list
list: ## Generate a complete list of available Makefile targets.
	@$(MAKE) -pRrq -f $(COMPONENT_DIR)/Makefile : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'

.PHONY: exec
exec:
	@docker run $(DOCKER_INTERACTIVE) \
    		-v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
    		$(DOCKER_CREATE_OPTS) bash

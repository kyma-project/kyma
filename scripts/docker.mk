# Default configuration
IMG_NAME := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG := $(DOCKER_TAG)
# BASE_PKG is a root packge of the component
BASE_PKG := github.com/kyma-project/kyma
# IMG_WSPPATH is a path to workspace in the container
IMG_WSPPATH := /workspace

# Other variables
# LOCAL_DIR in a local path to scripts folder
LOCAL_DIR = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
# COMPONENT_DIR is a local path to commponent
COMPONENT_DIR = $(shell pwd)
# WORKSPACE_LOCAL_DIR is a path to the scripts folder in the container
WORKSPACE_LOCAL_DIR = $(IMG_WSPPATH)/src/$(BASE_PKG)/scripts
# WORKSPACE_COMPONENT_DIR is a path to commponent in hte container
WORKSPACE_COMPONENT_DIR = $(IMG_WSPPATH)/src/$(BASE_PKG)/$(APP_PATH)

# Base docker configuration
DOCKER_CREATE_OPTS := -v $(LOCAL_DIR):$(WORKSPACE_LOCAL_DIR):delegated --rm -w $(WORKSPACE_COMPONENT_DIR) $(BUILDPACK)

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

.PHONY: release

release: build-image push-image

.PHONY: build-image push-image
build-image:
	docker build -t $(IMG_NAME) .
push-image:
	docker tag $(IMG_NAME) $(IMG_NAME):$(TAG)
	docker push $(IMG_NAME):$(TAG)
docker-create-opts:
	@echo $(DOCKER_CREATE_OPTS)


# Targets copying sources to buildpack
COPY_TARGETS = test
$(foreach t,$(COPY_TARGETS),$(eval $(call buildpack-cp-ro,$(t))))

.PHONY: list
list:
	@$(MAKE) -pRrq -f $(COMPONENT_DIR)/Makefile : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'

.PHONY: exec
exec:
	@docker run $(DOCKER_INTERACTIVE) \
    		-v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
    		$(DOCKER_CREATE_OPTS) bash
PKG = github.com/kyma-project/kyma/components/function-controller
ROOT :=  $(shell pwd)
# Image URL to use all building/pushing image targets
APP_NAME = function-controller
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20201113-a3e68400
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

override VERIFY_IGNORE := /vendor\|/automock\|/pkg/apis/serverless/v1alpha1/zz_generated.deepcopy.go

include $(SCRIPTS_DIR)/generic-make-go.mk

# Image URL to use all building/pushing image targets
IMG ?= $(APP_NAME):latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

pull-licenses-local:
	@echo "Omitting licenses will not be pulled."	

release:
	$(MAKE) gomod-release

# Run tests
test-local: export KUBEBUILDER_CONTROLPLANE_START_TIMEOUT = 2m
test-local: export KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT = 2m
test-local:
	go test -race -coverprofile=/tmp/artifacts/cover.out ./...
	@echo -n "Total coverage: "
	@go tool cover -func=/tmp/artifacts/cover.out | grep total | awk '{print $$3}'

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: controller-gen-local generate-local
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=serverless webhook paths="./..." \
		object:headerFile=hack/boilerplate.go.txt \
		output:crd:artifacts:config=config/crd/bases \
		output:rbac:artifacts:config=config/rbac \
		output:webhook:artifacts:config=config/webhook

CONTROLLER_GEN := $(shell which controller-gen)

# Generate code
.PHONY: generate-local
generate-local: controller-gen-local
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

.PHONY: build-image 
build-image: build-function-webhook \
	build-function-build-init \
	build-image-function-controller
	@echo "Override generic makefile build-image target to not execute"

.PHONY: push-image
push-image: push-function-webhook \
	push-function-build-init \
	push-image-function-controller
	@echo "Override generic makefile push-image target to not execute"

.PHONY: build-image push-image
build-image: build-function-webhook \
	build-function-build-init \
	build-image-function-controller
	@echo "Override generic makefile build-image target to not execute"


push-image: push-function-webhook \
	push-function-build-init \
	push-image-function-controller
	@echo "Override generic makefile push-image target to not execute"

######## function manager

MANAGER_NAME = function-controller
MANAGER_IMG_TAG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(MANAGER_NAME)

.PHONY: build-image-function-controller push-image-function-controller
build-image-function-controller:
	docker build -t $(MANAGER_NAME) -f $(ROOT)/deploy/manager/Dockerfile .

push-image-function-controller:
	docker tag $(MANAGER_NAME) $(MANAGER_IMG_TAG):$(DOCKER_TAG)
	docker push $(MANAGER_IMG_TAG):$(DOCKER_TAG)

######## function webhook

WEBHOOK_NAME = function-webhook
WEBHOOK_IMG_NAME = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(WEBHOOK_NAME)

.PHONY: build-function-webhook push-function-webhook
build-function-webhook:
	docker build -t $(WEBHOOK_NAME) -f $(ROOT)/deploy/webhook/Dockerfile .

push-function-webhook:
	docker tag $(WEBHOOK_NAME) $(WEBHOOK_IMG_NAME):$(DOCKER_TAG)
	docker push $(WEBHOOK_IMG_NAME):$(DOCKER_TAG)

######## builder init container image

JOBINIT_NAME = function-build-init
JOBINIT_IMG_NAME = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(JOBINIT_NAME)

.PHONY: build-function-build-init push-function-build-init
build-function-build-init:
	docker build -t $(JOBINIT_NAME) -f $(ROOT)/deploy/jobinit/Dockerfile .

push-function-build-init:
	docker tag $(JOBINIT_NAME) $(JOBINIT_IMG_NAME):$(DOCKER_TAG)
	docker push $(JOBINIT_IMG_NAME):$(DOCKER_TAG)

$(eval $(call buildpack-mount,resolve))
$(eval $(call buildpack-mount,mod-verify))
$(eval $(call buildpack-mount,controller-gen))
$(eval $(call buildpack-mount,generate))
$(eval $(call buildpack-mount,test))
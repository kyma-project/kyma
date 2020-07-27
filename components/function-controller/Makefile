PKG = github.com/kyma-project/kyma/components/function-controller
ROOT :=  $(shell pwd)
# Image URL to use all building/pushing image targets
APP_NAME = function-controller
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20200124-69faeef6
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

override VERIFY_IGNORE := /vendor\|/automock\|/pkg/apis/serverless/v1alpha1/zz_generated.deepcopy.go

include $(SCRIPTS_DIR)/generic-make-go.mk


# Image URL to use all building/pushing image targets
IMG ?= $(APP_NAME):latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

verify:: mod-verify
release: build-function-runtime-nodejs12-image \
	build-function-runtime-nodejs10-image \
	build-function-runtime-python38-image \
	push-function-runtime-nodejs12-image \
	push-function-runtime-nodejs10-image \
	push-function-runtime-python38-image \
	build-function-webhook \
	push-function-webhook \
	build-function-build-init \
	push-function-build-init \
	build-image push-image


resolve-local:
	GO111MODULE=on go mod download
	GO111MODULE=on go mod vendor

ensure-local:
	@echo "Go modules present in component - omitting."

dep-status:
	@echo "Go modules present in component - omitting."

dep-status-local:
	@echo "Go modules present in component - omitting."

mod-verify-local:
	GO111MODULE=on go mod verify

# Run tests
test-local: export KUBEBUILDER_CONTROLPLANE_START_TIMEOUT = 2m
test-local: export KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT = 2m
test-local: generate-local manifests
	GO111MODULE=on go test -v -race -coverprofile=/tmp/artifacts/cover.out -mod vendor ./...
	@echo -n "Total coverage: "
	@go tool cover -func=/tmp/artifacts/cover.out | grep total | awk '{print $$3}'

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

undeploy: uninstall
	kustomize build config/default | kubectl delete  -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: install manifests
	kubectl apply -k config/default

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: controller-gen-local generate-local
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=serverless webhook paths="./..." \
		object:headerFile=hack/boilerplate.go.txt \
		output:crd:artifacts:config=config/crd/bases \
		output:rbac:artifacts:config=config/rbac \
		output:webhook:artifacts:config=config/webhook

# Generate code
.PHONY: generate-local
generate-local: controller-gen-local
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# find or download controller-gen
# download controller-gen if necessary
.PHONY: controller-gen-local
controller-gen-local:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.4 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/serverless"

######## function manager

MANAGER_NAME = function-controller
MANAGER_IMG_TAG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(MANAGER_NAME)

.PHONY: build-image push-image
build-image:
	docker build -t $(MANAGER_NAME) -f $(ROOT)/deploy/manager/Dockerfile .

push-image:
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

######## function runtime image

## nodejs12
FUNCTION_RUNTIME_IMAGE_NODEJS12 = function-runtime-nodejs12
FUNCTION_RUNTIME_NODEJS12_IMG_NAME=$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(FUNCTION_RUNTIME_IMAGE_NODEJS12)

.PHONY: build-function-runtime-nodejs12-image push-function-runtime-nodejs12-image
build-function-runtime-nodejs12-image:
	docker build -t $(FUNCTION_RUNTIME_IMAGE_NODEJS12) $(ROOT)/deploy/function-runtime-image/nodejs12/

push-function-runtime-nodejs12-image:
	docker tag $(FUNCTION_RUNTIME_IMAGE_NODEJS12) $(FUNCTION_RUNTIME_NODEJS12_IMG_NAME):$(DOCKER_TAG)
	docker push $(FUNCTION_RUNTIME_NODEJS12_IMG_NAME):$(DOCKER_TAG)

## nodejs10
FUNCTION_RUNTIME_IMAGE_NODEJS10 = function-runtime-nodejs10
FUNCTION_RUNTIME_NODEJS10_IMG_NAME=$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(FUNCTION_RUNTIME_IMAGE_NODEJS10)

.PHONY: build-function-runtime-nodejs10-image push-function-runtime-nodejs10-image
build-function-runtime-nodejs10-image:
	docker build -t $(FUNCTION_RUNTIME_IMAGE_NODEJS10) $(ROOT)/deploy/function-runtime-image/nodejs10/

push-function-runtime-nodejs10-image:
	docker tag $(FUNCTION_RUNTIME_IMAGE_NODEJS10) $(FUNCTION_RUNTIME_NODEJS10_IMG_NAME):$(DOCKER_TAG)
	docker push $(FUNCTION_RUNTIME_NODEJS10_IMG_NAME):$(DOCKER_TAG)

## python38
FUNCTION_RUNTIME_IMAGE_PYTHON38 = function-runtime-python38
FUNCTION_RUNTIME_PYTHON38_IMG_NAME=$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(FUNCTION_RUNTIME_IMAGE_PYTHON38)

.PHONY: build-function-runtime-python38-image push-function-runtime-python38-image
build-function-runtime-python38-image:
	docker build -t $(FUNCTION_RUNTIME_IMAGE_PYTHON38) $(ROOT)/deploy/function-runtime-image/python38/

push-function-runtime-python38-image:
	docker tag $(FUNCTION_RUNTIME_IMAGE_PYTHON38) $(FUNCTION_RUNTIME_PYTHON38_IMG_NAME):$(DOCKER_TAG)
	docker push $(FUNCTION_RUNTIME_PYTHON38_IMG_NAME):$(DOCKER_TAG)

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
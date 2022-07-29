PKG = github.com/kyma-project/kyma/components/function-controller
ROOT :=  $(shell pwd)
# Image URL to use all building/pushing image targets
APP_NAME = function-controller
APP_PATH = components/$(APP_NAME)
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

override VERIFY_IGNORE := /vendor\|/automock\|/pkg/apis/serverless/v1alpha2/zz_generated.deepcopy.go

include $(SCRIPTS_DIR)/generic-make-go.mk

# Image URL to use all building/pushing image targets
IMG ?= $(APP_NAME):latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,crdVersions=v1"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

release:
	$(MAKE) gomod-release-local

# Run tests
test-local: export KUBEBUILDER_CONTROLPLANE_START_TIMEOUT = 2m
test-local: export KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT = 2m
test-local:
	mkdir -p /tmp/artifacts
	go test -count=1 -race -coverprofile=/tmp/artifacts/cover.out ./...
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
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.9 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

.PHONY: build-image 
build-image: build-image-function-webhook \
	build-image-function-build-init \
	build-image-function-controller
	@echo "Override generic makefile build-image target to not execute"

.PHONY: push-image
push-image: push-function-webhook \
	push-function-build-init \
	push-image-function-controller
	@echo "Override generic makefile push-image target to not execute"

.PHONY: build-image push-image
build-image: build-image-function-webhook \
	build-image-function-build-init \
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
ifeq ($(JOB_TYPE), postsubmit)
	@echo "Sign image with Cosign"
	cosign version
	cosign sign -key ${KMS_KEY_URL} $(MANAGER_IMG_TAG):$(DOCKER_TAG)
else
	@echo "Image signing skipped"
endif

install-manager-k3d: build-image-function-controller
	$(eval HASH_TAG=$(shell docker images $(MANAGER_NAME):latest --quiet))
	docker tag $(MANAGER_NAME) $(MANAGER_NAME):$(HASH_TAG)

	k3d image import $(MANAGER_NAME):$(HASH_TAG) -c kyma
	kubectl set image deployment serverless-ctrl-mngr -n kyma-system manager=$(MANAGER_NAME):$(HASH_TAG)

######## function webhook

WEBHOOK_NAME = function-webhook
WEBHOOK_IMG_NAME = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(WEBHOOK_NAME)

.PHONY: build-image-function-webhook push-function-webhook
build-image-function-webhook:
	docker build -t $(WEBHOOK_NAME) -f $(ROOT)/deploy/webhook/Dockerfile .

push-function-webhook:
	docker tag $(WEBHOOK_NAME) $(WEBHOOK_IMG_NAME):$(DOCKER_TAG)
	docker push $(WEBHOOK_IMG_NAME):$(DOCKER_TAG)
ifeq ($(JOB_TYPE), postsubmit)
	@echo "Sign image with Cosign"
	cosign version
	cosign sign -key ${KMS_KEY_URL} $(WEBHOOK_IMG_NAME):$(DOCKER_TAG)
else
	@echo "Image signing skipped"
endif

install-webhook-k3d: build-image-function-webhook
	$(eval HASH_TAG=$(shell docker images $(WEBHOOK_NAME):latest --quiet))
	docker tag $(WEBHOOK_NAME) $(WEBHOOK_NAME):$(HASH_TAG)

	k3d image import $(WEBHOOK_NAME):$(HASH_TAG) -c kyma
	kubectl set image deployment serverless-webhook-svc -n kyma-system webhook=$(WEBHOOK_NAME):$(HASH_TAG)
######## builder init container image

JOBINIT_NAME = function-build-init
JOBINIT_IMG_NAME = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(JOBINIT_NAME)

.PHONY: build-image-function-build-init push-function-build-init
build-image-function-build-init:
	docker build -t $(JOBINIT_NAME) -f $(ROOT)/deploy/jobinit/Dockerfile .

push-function-build-init:
	docker tag $(JOBINIT_NAME) $(JOBINIT_IMG_NAME):$(DOCKER_TAG)
	docker push $(JOBINIT_IMG_NAME):$(DOCKER_TAG)
ifeq ($(JOB_TYPE), postsubmit)
	@echo "Sign image with Cosign"
	cosign version
	cosign sign -key ${KMS_KEY_URL} $(JOBINIT_IMG_NAME):$(DOCKER_TAG)
else
	@echo "Image signing skipped"
endif

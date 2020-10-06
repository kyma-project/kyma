APP_NAME = eventing-controller
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20200930-31e9b318
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

override ENTRYPOINT = main.go

include $(SCRIPTS_DIR)/generic-make-go.mk

# DEV_IMG is the image used during development workflow only
DEV_IMG = ko://github.com/kyma-project/kyma/components/eventing-controller
NAMESPACE = "eventing"
VERIFY_IGNORE := /vendor\|/mocks
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:crdVersions=v1"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifneq (,$(shell which go))
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif
endif

all: manager

# Build controller binary
manager: generate-local fmt-local vet-local
	go build -o bin/controller main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate-local fmt-local vet-local manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | ko apply -f -

# Install sample CRs into a cluster
install-samples: manifests
	ko apply -f config/samples

# Uninstall sample CRs into a cluster
uninstall-samples: manifests
	ko delete -f config/samples

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | ko delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config for local development
deploy-local: manifests deploy-local-dry-run
	kubectl create ns ${NAMESPACE} || true
	kustomize build config/default | ko apply -f -

undeploy-local: manifests
	kustomize build config/default | kubectl delete -f -

# Dry run for deploying controller for local development
deploy-local-dry-run: manifests
	cd config/manager && kustomize edit set image controller=${DEV_IMG} && kustomize edit set namespace ${NAMESPACE}
	cd config/rbac && kustomize edit set namespace ${NAMESPACE}
	kustomize build config/default

# Generate manifests e.g. CRD, RBAC etc.
manifests-local: controller-gen-local
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Generate code
generate-local: controller-gen-local
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

verify:: mod-verify

resolve-local:
	GO111MODULE=on go mod vendor -v

ensure-local:
	@echo "Go modules present in component - omitting."

dep-status:
	@echo "Go modules present in component - omitting."

dep-status-local:
	@echo "Go modules present in component - omitting."

mod-verify-local:
	GO111MODULE=on go mod verify

test-local: generate-local fmt-local vet-local
	go test ./... -coverprofile cover.out

release: resolve verify build-image push-image

path-to-referenced-charts:
	@echo "resources/eventing-manager/eventing-controller"

# find or download controller-gen
# download controller-gen if necessary
.PHONY: controller-gen-local
controller-gen-local:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	if [ ! $(which go) ]; then \
		go mod init tmp ; \
		go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ; \
		rm -rf $$CONTROLLER_GEN_TMP_DIR ; \
	fi; \
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

$(eval $(call buildpack-cp-ro,resolve))
$(eval $(call buildpack-mount,mod-verify))
$(eval $(call buildpack-mount,test))
$(eval $(call buildpack-mount,controller-gen))
$(eval $(call buildpack-mount,generate))
$(eval $(call buildpack-mount,manifests))
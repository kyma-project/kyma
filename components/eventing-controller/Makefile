APP_NAME = eventing-controller
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20210601-24c60b5a
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

override ENTRYPOINT = main.go

include $(SCRIPTS_DIR)/generic-make-go.mk

##PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
PROJECT_DIR := $(shell pwd)

NAMESPACE = "kyma-system"
VERIFY_IGNORE := /vendor\|/mocks

# Image URL to use all building/pushing image targets
IMG ?= eventing-controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

GOBUILD_FLAGS := -v
PKG := github.com/kyma-project/kyma/components/eventing-controller

##@ General

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
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
# Deploy eventing-controller in the configured Kubernetes cluster in ~/.kube/config for local development
deploy: manifests-local set-up-local-env ## build the binary locally using `ko` and deploy it to a cluster
	kustomize build config/default | envsubst | ko apply -f -

manifests-local: controller-gen-local ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate-local: controller-gen-local ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

set-up-local-env:
    export NATS_URL ?= eventing-nats.kyma-system.svc.cluster.local
    export EVENT_TYPE_PREFIX ?= sap.kyma.custom
    export DOMAIN ?= example-domain.com
    export WEBHOOK_TOKEN_ENDPOINT ?= "https://oauth2.${DOMAIN}/oauth2/token"


ENVTEST_ASSETS_DIR=$(PROJECT_DIR)/testbin/$(shell uname)
test-local: test-local-nonlocal; ## Run tests.

setup-envtest:
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR)

test-local-nonlocal: manifests-local generate-local setup-envtest
	go test ./... -coverprofile cover.out

test-local-local: manifests-local generate-local setup-envtest
	go test ./... -coverprofile cover.out --tags=local

check-code: check-imports fmt-local vet-local lint ## Run various linters and other code checks. Use before committing

run: manifests-local generate-local fmt-local vet-local set-up-local-env## Run a controller from your host. Runs with buildtags `local`
	go run --tags=local ./cmd/eventing-controller/main.go

eventing-controller: gomod-vendor-local
        @CGO_ENABLED=0 go build -o bin/eventing-controller \
        $(GOBUILD_FLAGS) \
        $(PKG)/cmd/eventing-controller


##@ Release
##release: resolve-local generate-local verify-local build-image push-image copy-external-crds ## Job used by our CI to release a binary
release:
	$(MAKE) gomod-release-local

##@ Deployment

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

CONTROLLER_GEN = $(PROJECT_DIR)/bin/$(shell uname)/controller-gen
controller-gen-local: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

path-to-referenced-charts:
	@echo "resources/eventing"


deploy-local-dry-run: set-up-local-env manifests-local ## Dry run for deploying eventing-controller for local development
	cd config/rbac && kustomize edit set namespace ${NAMESPACE}
	kustomize build config/default | envsubst

##@ CRD

copy-crds: manifests-local ## copy Eventing CRDs to installation/eventing
	cp ./config/crd/bases/eventing.kyma-project.io_subscriptions.yaml ./../../installation/resources/crds/eventing/subscriptions.eventing.kyma-project.io.crd.yaml
	cp ./config/crd/bases/eventing.kyma-project.io_eventingbackends.yaml ./../../installation/resources/crds/eventing/eventingbackends.eventing.kyma-project.io.crd.yaml

copy-external-crds: ## copy external CRDs to config/crd/external
	mkdir -p config/crd/external
	cp ../../installation/resources/crds/api-gateway/apirules.gateway.crd.yaml config/crd/external/apirules-gateway-kyma-project-io.yaml
	cp ../../installation/resources/crds/application-connector/applications.applicationconnector.crd.yaml config/crd/external/applications-applicationconnector-kyma-project-io.yaml

# go-get-tool will 'go get' any package $2 and install it to $1.
ARCH := $(shell uname)
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin/$(ARCH) go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

$(eval $(call buildpack-cp-ro,resolve))
$(eval $(call buildpack-mount,mod-verify))
$(eval $(call buildpack-mount,test))
$(eval $(call buildpack-mount,controller-gen))
$(eval $(call buildpack-mount,generate))
$(eval $(call buildpack-mount,manifests))

resolve_clean:
	rm -rf vendor

licenses_clean:
	rm -rf licenses

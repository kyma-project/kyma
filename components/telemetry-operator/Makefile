
APP_NAME = telemetry-operator
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20220428-6e81d2c4
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
PROJECT_DIR := $(shell pwd)
OS := $(shell uname)

include $(SCRIPTS_DIR)/generic-make-go.mk

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.21

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifneq (,$(shell which go))
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

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

# Generate manifests e.g. CRD, RBAC etc.
manifests-local: controller-gen-local
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

copy-crds-local: manifests-local
	@cp ./config/crd/bases/telemetry.kyma-project.io_logpipelines.yaml ./../../installation/resources/crds/telemetry/logpipelines.crd.yaml

generate-local: controller-gen-local
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

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

 ## Run tests.
test-local: manifests-local generate-local fmt-local vet-local envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out

build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

lint: ## Run various linters
	../../hack/verify-lint.sh .

tls.key:
	@openssl genrsa -out tls.key 4096

tls.crt: tls.key
	@openssl req -sha256 -new -key tls.key -out tls.csr -subj '/CN=localhost'
	@openssl x509 -req -sha256 -days 3650 -in tls.csr -signkey tls.key -out tls.crt
	@rm tls.csr

gen-webhook-cert-local: tls.key tls.crt

run-local: gen-webhook-cert-local manifests-local generate-local fmt-local vet-local ## Run a controller from your host.
	go run ./main.go --cm-name=telemetry-fluent-bit --sections-cm-name=telemetry-fluent-bit-sections --parser-cm-name=telemetry-fluent-bit-parsers --ds-name=telemetry-fluent-bit --env-secret=telemetry-fluent-bit-env --files-cm=telemetry-fluent-bit-files --fluent-bit-ns=kyma-system --cert-dir=.

## Will be called from Prow-Pipeline; using targets from generic make file
release: resolve generate verify build-image push-image

##@ Deployment

install-crds-local: manifests-local ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kustomize build config/crd | kubectl apply -f -

uninstall-crds-local: manifests-local ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	kustomize build config/crd | kubectl delete -f -

ENVTEST = $(PROJECT_DIR)/bin/$(OS)/setup-envtest
.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

CONTROLLER_GEN = $(PROJECT_DIR)/bin/$(OS)/controller-gen
.PHONY: controller-gen-local
controller-gen-local: ## find or download controller-gen if necessary
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.0)

# go-get-tool will 'go get' any package $2 and install it to $1.
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin/$(OS) go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

##@ Dynamic Function Build
$(eval $(call buildpack-cp-ro,resolve))
$(eval $(call buildpack-mount,mod-verify))
$(eval $(call buildpack-mount,test))
$(eval $(call buildpack-mount,controller-gen))
$(eval $(call buildpack-mount,manifests))

APP_NAME = eventing-controller
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20220407-4da6c929
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

override ENTRYPOINT = cmd/eventing-controller/main.go
override IGNORE_LINTING_ISSUES =
override ADDITIONAL_LINTERS="gocognit"

include $(SCRIPTS_DIR)/generic-make-go.mk

PROJECT_DIR := $(shell pwd)

NAMESPACE = "kyma-system"
VERIFY_IGNORE := /vendor\|/mocks

# Image URL to use all building/pushing image targets
IMG ?= eventing-controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
#
CRD_OPTIONS ?= "crd"
MANIFEST_PATH ?= "./..."

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

CONTROLLER_TOOLS_VERSION ?= v0.11.1

help-generic:
	@printf "\nUsage:\n  make \033[36m<target>\033[0m\n"
	@printf "\nTargets from common makefile:\n"

##@ Development
HELM_TEMPLATE_ASSETS_DIR=./config/helm/manifests
HELM_TEMPLATE_CRD_PATCHES_DIR=./config/helm/localcrd
# Deploy eventing-controller in the configured Kubernetes cluster in ~/.kube/config for local development using helm templates.
deploy: manifests-local copy-crds copy-external-crds set-up-local-env ## build the binary locally using `ko` and deploy it to a cluster
	$(shell cd ../event-publisher-proxy; ko publish github.com/kyma-project/kyma/components/event-publisher-proxy/cmd/event-publisher-proxy -B > /dev/null)
	rm -rf ${HELM_TEMPLATE_ASSETS_DIR}; mkdir -p ${HELM_TEMPLATE_ASSETS_DIR}
	helm template --output-dir ${HELM_TEMPLATE_ASSETS_DIR} eventing ../../resources/eventing
	kustomize build config/helm | envsubst | ko apply -f -
	rm -rf ${HELM_TEMPLATE_ASSETS_DIR}

deploy-dry-run: manifests-local copy-crds copy-external-crds set-up-local-env ## Dry run for deploying eventing-controller for local development using helm templates.
	$(shell cd ../event-publisher-proxy; ko publish github.com/kyma-project/kyma/components/event-publisher-proxy/cmd/event-publisher-proxy -B > /dev/null)
	rm -rf ${HELM_TEMPLATE_ASSETS_DIR}; mkdir -p ${HELM_TEMPLATE_ASSETS_DIR}
	helm template --output-dir ${HELM_TEMPLATE_ASSETS_DIR} eventing ../../resources/eventing
	kustomize build config/helm | envsubst
	rm -rf ${HELM_TEMPLATE_ASSETS_DIR}

manifests-local: controller-gen-local ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths=$(MANIFEST_PATH) output:crd:artifacts:config=config/crd/bases

generate-local: controller-gen-local ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

set-up-local-env:
	export NATS_URL ?= eventing-nats.kyma-system.svc.cluster.local
	export DOMAIN ?= example-domain.com
	export WEBHOOK_TOKEN_ENDPOINT ?= "https://oauth2.${DOMAIN}/oauth2/token"

ENVTEST_ASSETS_DIR=$(PROJECT_DIR)/testbin/$(shell uname)
test-local: go-test; ## Run tests.

ENVTEST ?= $(LOCALBIN)/setup-envtest

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: setup-envtest
setup-envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

go-test: manifests-local generate-local setup-envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

check-code: check-imports fmt-local vet-local lint ## Run various linters and other code checks. Use before committing

run: manifests-local generate-local fmt-local vet-local set-up-local-env## Run a controller from your host. Runs with buildtags `local`
	go run --tags=local ./cmd/eventing-controller/main.go

eventing-controller: gomod-vendor-local ## Build the binary
	@CGO_ENABLED=0 go build -o bin/$(APP_NAME) \
	$(GOBUILD_FLAGS) \
	$(PKG)/$(ENTRYPOINT)

test: copy-external-crds ## It will also run overridden `test` target from generic makefile after `copy-external-crds`

##@ Release

##release: resolve-local generate-local verify-local build-image push-image copy-external-crds ## Job used by our CI to release a binary
release:
	$(MAKE) gomod-release-local

##@ Deployment

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	rm -rf ${HELM_TEMPLATE_ASSETS_DIR}; mkdir -p ${HELM_TEMPLATE_ASSETS_DIR}
	helm template --output-dir ${HELM_TEMPLATE_ASSETS_DIR} eventing ../../resources/eventing
	kustomize build config/helm | envsubst | kubectl delete -f -
	rm -rf ${HELM_TEMPLATE_ASSETS_DIR}

CONTROLLER_GEN = $(PROJECT_DIR)/bin/$(shell uname)/controller-gen
OS := $(shell uname)
controller-gen-local: ## Download controller-gen locally if necessary.
	GOBIN=$(PROJECT_DIR)/bin/$(OS) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

##@ CRD

copy-crds: manifests-local ## copy Eventing CRDs to installation/eventing
	rm -rf $(HELM_TEMPLATE_CRD_PATCHES_DIR); mkdir $(HELM_TEMPLATE_CRD_PATCHES_DIR); kustomize build config/crd -o $(HELM_TEMPLATE_CRD_PATCHES_DIR);
	cp $(HELM_TEMPLATE_CRD_PATCHES_DIR)/apiextensions.k8s.io_v1_customresourcedefinition_subscriptions.eventing.kyma-project.io.yaml ./../../installation/resources/crds/eventing/subscriptions.eventing.kyma-project.io.crd.yaml
	cp ./config/crd/bases/eventing.kyma-project.io_eventingbackends.yaml ./../../installation/resources/crds/eventing/eventingbackends.eventing.kyma-project.io.crd.yaml

copy-external-crds: ## copy external CRDs to config/crd/external
	mkdir -p config/crd/external
	cp ../../installation/resources/crds/api-gateway/apirules.gateway.crd.yaml config/crd/external/apirules-gateway-kyma-project-io.yaml

$(eval $(call buildpack-cp-ro,resolve))
$(eval $(call buildpack-mount,mod-verify))
$(eval $(call buildpack-mount,test))
$(eval $(call buildpack-mount,controller-gen))
$(eval $(call buildpack-mount,generate))
$(eval $(call buildpack-mount,manifests))

resolve_clean:
	rm -rf vendor

lint-thoroughly:
	golangci-lint run

update_metrics_docs:
	@promlinter list -ojson ../event-publisher-proxy > metrics_epp.json
	@jq 'sort_by(.Name)'  metrics_epp.json > metrics_epp_sorted.json
	@promlinter list -ojson ../eventing-controller/ > metrics_ec.json
	@jq 'sort_by(.Name)'  metrics_ec.json > metrics_ec_sorted.json
	@gomplate -d ec=metrics_ec_sorted.json -d epp=metrics_epp_sorted.json -f hack/metrics.doc.tpl  | prettier --parser markdown > ../../docs/04-operation-guides/operations/evnt-02-eventing-metrics.md

update_docs: copy-crds
	go run ../../hack/table-gen/main.go --crd-filename ../../installation/resources/crds/eventing/subscriptions.eventing.kyma-project.io.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/evnt-01-subscription.md
	go run ../../hack/table-gen/main.go --crd-filename ../../installation/resources/crds/eventing/eventingbackends.eventing.kyma-project.io.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/evnt-02-eventingbackend.md

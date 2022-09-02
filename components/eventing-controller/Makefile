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
CRD_OPTIONS ?= "crd"

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

help-generic:
	@printf "\nUsage:\n  make \033[36m<target>\033[0m\n"
	@printf "\nTargets from common makefile:\n"

##@ Development
HELM_TEMPLATE_ASSETS_DIR=./config/helm/manifests
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
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate-local: controller-gen-local ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

set-up-local-env:
    export NATS_URL ?= eventing-nats.kyma-system.svc.cluster.local
    export DOMAIN ?= example-domain.com
    export WEBHOOK_TOKEN_ENDPOINT ?= "https://oauth2.${DOMAIN}/oauth2/token"

ENVTEST_ASSETS_DIR=$(PROJECT_DIR)/testbin/$(shell uname)
test-local: go-test; ## Run tests.

setup-envtest:
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR)

go-test: manifests-local generate-local setup-envtest
	KUBEBUILDER_ASSETS=${ENVTEST_ASSETS_DIR}/bin go test ./... -coverprofile cover.out

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
	GOBIN=$(PROJECT_DIR)/bin/$(OS) go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1

path-to-referenced-charts:
	@echo "resources/eventing"

##@ CRD

copy-crds: manifests-local ## copy Eventing CRDs to installation/eventing
	rm -rf localcopy; mkdir localcopy; kustomize build config/crd -o localcopy
	cp ./localcopy/apiextensions.k8s.io_v1_customresourcedefinition_subscriptions.eventing.kyma-project.io.yaml ./../../installation/resources/crds/eventing/subscriptions.eventing.kyma-project.io.crd.yaml
	cp ./config/crd/bases/eventing.kyma-project.io_eventingbackends.yaml ./../../installation/resources/crds/eventing/eventingbackends.eventing.kyma-project.io.crd.yaml
	rm -rf localcopy
copy-external-crds: ## copy external CRDs to config/crd/external
	mkdir -p config/crd/external
	cp ../../installation/resources/crds/api-gateway/apirules.gateway.crd.yaml config/crd/external/apirules-gateway-kyma-project-io.yaml
	cp ../../installation/resources/crds/application-connector/applications.applicationconnector.crd.yaml config/crd/external/applications-applicationconnector-kyma-project-io.yaml

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

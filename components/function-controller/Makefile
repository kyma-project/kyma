APP_NAME = function-controller
APP_PATH = components/function-controller
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder:v20190208-813daef
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
PKG = github.com/kyma-project/kyma/components/function-controller

include $(SCRIPTS_DIR)/go-dep.mk

MAIN_PATH := ./cmd/manager/main.go

all: test manager

# Run tests
# Default is 20s - available since controller-runtime 0.1.5
test-local: export KUBEBUILDER_CONTROLPLANE_START_TIMEOUT = 2m
# Default is 20s - available since controller-runtime 0.1.5
test-local: export KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT = 2m
test-local: generate fmt vet
	go test -race -v ./pkg/... ./cmd/... -coverprofile cover.out
	@echo "Total test coverage: $$(go tool cover -func=cover.out | grep total | awk '{print $$3}')"

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager $(PKG)/cmd/manager

# NOTE: does not work anymore until https://github.com/kubernetes-sigs/kubebuilder/issues/400 is implemented
# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install:
	kubectl apply -f config/crds/

# Uninstall CRDs from cluster
uninstall:
	kubectl delete -f config/crds/

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

deploy: manifests install bin/kustomize ko
	bin/kustomize build config/default | ko apply -f -

removewebhook:
	-kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io function-webhook.serverless.kyma-project.io

undeploy: uninstall removewebhook bin/kustomize ko
	bin/kustomize build config/default | ko delete  -f -

# used by 'go install'
GOPATH ?= $(shell go env GOPATH)
export GOBIN := $(GOPATH)/src/$(PKG)/bin

.PHONY: bin/kustomize
bin/kustomize:
	@go install ./vendor/sigs.k8s.io/kustomize

.PHONY: ko
ko:
	@go get -u github.com/google/ko/cmd/ko

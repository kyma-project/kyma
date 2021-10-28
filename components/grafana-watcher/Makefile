
APP_NAME = grafana-watcher
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20211013-2c064d85
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

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


resolve-local:
	GO111MODULE=on go mod vendor -v
ensure-local:
	@echo "Go modules present in component - omitting."

dep-status:
	@echo "Go modules present in component - omitting."

dep-status-local:
	@echo "Go modules present in component - omitting."

build: generate fmt vet ## Build manager binary.
	go build -o bin/grafana-watcher main.go

## Will be called from Prow-Pipeline; using targets from generic make file
release: verify build-image push-image

##@ Dynamic Function Build
$(eval $(call buildpack-cp-ro,resolve))
$(eval $(call buildpack-mount,mod-verify))
$(eval $(call buildpack-mount,test))
$(eval $(call buildpack-mount,controller-gen))
$(eval $(call buildpack-mount,manifests))
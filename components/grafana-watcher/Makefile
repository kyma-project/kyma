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

resolve-local:
	GO111MODULE=on go mod vendor -v

## Will be called from Prow-Pipeline; using targets from generic make file
release: resolve generate verify build-image push-image

##@ Dynamic Function Build
$(eval $(call buildpack-cp-ro,resolve))
APP_NAME = permission-controller
APP_PATH = components/$(APP_NAME)
# todo: kubebuilder2 buildpack should be based on toolbox
# BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20191011-51ed45a
BUILDPACK = jakkab/kubebuilder:0.0.1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

override ENTRYPOINT := main.go

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

include $(SCRIPTS_DIR)/generic-make-go.mk

verify:: mod-verify

# Skip dep commands

ensure-local:
	@echo "Go modules present in component - omitting."

dep-status:
	@echo "Go modules present in component - omitting."

dep-status-local:
	@echo "Go modules present in component - omitting."

# go mod

resolve-local:
	GO111MODULE=on go mod vendor -v

mod-verify-local:
	GO111MODULE=on go mod verify

test-local:
	GO111MODULE=on go test ./...

$(eval $(call buildpack-mount,resolve))
$(eval $(call buildpack-mount,mod-verify))
$(eval $(call buildpack-mount,test))

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/permission-controller"


###############
# development #
###############

ifndef EXCLUDED_NAMESPACES
override EXCLUDED_NAMESPACES = "kyma-system, istio-system, default, knative-eventing, knative-serving, kube-node-lease, kube-public, kube-system, kyma-installer, kyma-integration, natss"
endif

ifndef SUBJECT_GROUPS
override SUBJECT_GROUPS = "example, groups, replace, me"
endif

ifndef STATIC_CONNECTOR
override STATIC_CONNECTOR = true
endif

.PHONY: run
run:
	go run main.go --excluded-namespaces=${EXCLUDED_NAMESPACES} --subject-groups=${SUBJECT_GROUPS} --static-connector=${STATIC_CONNECTOR}






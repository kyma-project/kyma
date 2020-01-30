APP_NAME = permission-controller
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20200114-1bab4ab8
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
override ENTRYPOINT := main.go

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

EXCLUDED_NAMESPACES ?= "kyma-system, istio-system, default, knative-eventing, knative-serving, kube-node-lease, kube-public, kube-system, kyma-installer, kyma-integration, natss"
SUBJECT_GROUPS ?= "namespace-admins"
STATIC_CONNECTOR ?= true

.PHONY: run
run:
	go run main.go --excluded-namespaces=${EXCLUDED_NAMESPACES} --subject-groups=${SUBJECT_GROUPS} --static-connector=${STATIC_CONNECTOR}






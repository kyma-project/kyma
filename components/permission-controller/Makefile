APP_NAME = permission-controller
APP_PATH = components/$(APP_NAME)
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
override ENTRYPOINT := main.go

include $(SCRIPTS_DIR)/generic-make-go.mk

release:
	$(MAKE) gomod-release-local

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/permission-controller"

###############
# development #
###############

EXCLUDED_NAMESPACES ?= "kyma-system, istio-system, default, knative-eventing, kube-node-lease, kube-public, kube-system, kyma-installer, kyma-integration, natss, tekton-pipelines"
SUBJECT_GROUPS ?= "runtimeNamespaceAdmin"
STATIC_CONNECTOR ?= true

.PHONY: run
run:
	go run main.go --excluded-namespaces=${EXCLUDED_NAMESPACES} --subject-groups=${SUBJECT_GROUPS} --static-connector=${STATIC_CONNECTOR}

APP_NAME = console-backend-service
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/go-modules.mk

VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg


.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

dev-deps:
	go get github.com/githubnemo/CompileDaemon

watch: export APP_KUBECONFIG_PATH=/Users/${USER}/.kube/config
watch: export APP_VERBOSE=true
watch: export APP_RAFTER_ADDRESS=https://storage.${KYMA_DOMAIN}
watch: export APP_RAFTER_VERIFY_SSL=false
watch: export APP_APPLICATION_GATEWAY_INTEGRATION_NAMESPACE=kyma-integration
watch: export APP_APPLICATION_CONNECTOR_URL=http://dummy.url
watch: export APP_OIDC_ISSUER_URL=https://dex.${KYMA_DOMAIN}
watch: export APP_OIDC_CLIENT_ID=kyma-client

watch:
	CompileDaemon -build="make build-local" -command="go run main.go"
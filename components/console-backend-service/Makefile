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

run: export APP_KUBECONFIG_PATH=/Users/${USER}/.kube/config
run: export APP_VERBOSE=true
run: export APP_RAFTER_ADDRESS=https://storage.${KYMA_DOMAIN}
run: export APP_RAFTER_VERIFY_SSL=false
run: export APP_APPLICATION_GATEWAY_INTEGRATION_NAMESPACE=kyma-integration
run: export APP_APPLICATION_CONNECTOR_URL=http://dummy.url
run: export APP_OIDC_ISSUER_URL=https://dex.${KYMA_DOMAIN}
run: export APP_OIDC_CLIENT_ID=kyma-client

run:
	./console-backend-service

watch:
	CompileDaemon -build="make build-local" -command="go run main.go" -graceful-kill

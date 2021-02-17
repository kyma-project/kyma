APP_NAME = application-gateway
APP_PATH = components/$(APP_NAME)
# BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20200423-1d9d6590
SCRIPTS_DIR = $(realpath $(shell pwd))

override ENTRYPOINT = cmd/applicationgateway/

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

# think this might work
release: gomod-release

# verify:: gomod-verify-local

# resolve-local:
# 	echo "$$(date +"%Y/%m/%d %T %Z") vendor"
# 	GOPROXY=http://athens-proxy.c.sap-kyma-prow.internal:80 GO111MODULE=on go mod vendor
# 	echo "$$(date +"%Y/%m/%d %T %Z") vendor done"

# ensure-local:
# 	@echo "Go modules present in component - omitting."

# dep-status:
# 	@echo "Go modules present in component - omitting."

# dep-status-local:
# 	@echo "Go modules present in component - omitting."

# test-local:
# 	echo "$$(date +"%Y/%m/%d %T %Z") test"
# 	GOPROXY=http://athens-proxy.c.sap-kyma-prow.internal:80 GO111MODULE=on go test ./...
# 	echo "$$(date +"%Y/%m/%d %T %Z") test done"

# $(eval $(call buildpack-cp-ro,resolve))
# $(eval $(call buildpack-mount,mod-verify))
# $(eval $(call buildpack-mount,test))

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/application-connector"

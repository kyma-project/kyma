APP_NAME = apiserver-proxy
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
include $(SCRIPTS_DIR)/generic-make-go.mk

ENTRYPOINT := ./cmd/proxy/main.go

verify:: mod-verify


.PHONY: clean
clean:
	rm -rf bin

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/core"

ensure-local:
	@echo "Go modules present in component - omitting."

dep-status:
	@echo "Go modules present in component - omitting."

dep-status-local:
	@echo "Go modules present in component - omitting."

resolve-local:
	GO111MODULE=on go mod vendor -v

test-local:
	GO111MODULE=on go test ./...

mod-verify-local:
	GO111MODULE=on go mod verify

$(eval $(call buildpack-cp-ro,resolve))
$(eval $(call buildpack-mount,mod-verify))
$(eval $(call buildpack-mount,test))

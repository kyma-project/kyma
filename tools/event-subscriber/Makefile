APP_NAME = event-subscriber-tools
APP_PATH = tools/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.14

SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

override ENTRYPOINT = cmd/

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

verify:: mod-verify

resolve-local:
	GO111MODULE=on go mod vendor -v

ensure-local:
	@echo "Go modules present in component - omitting."

dep-status:
	@echo "Go modules present in component - omitting."

dep-status-local:
	@echo "Go modules present in component - omitting."

mod-verify-local:
	GO111MODULE=on go mod verify 

test-local:
	GO111MODULE=on go test ./... 

$(eval $(call buildpack-cp-ro,resolve)) 
$(eval $(call buildpack-mount,mod-verify)) 
$(eval $(call buildpack-mount,test))

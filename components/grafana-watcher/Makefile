
APP_NAME = grafana-watcher
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20210723-91c9654f
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

include $(SCRIPTS_DIR)/generic-make-go.mk

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

build: generate fmt vet ## Build manager binary.
	go build -o bin/grafana-watcher main.go

## Will be called from Prow-Pipeline; using targets from generic make file
release: resolve generate verify build-image push-image

##@ Dynamic Function Build
$(eval $(call buildpack-cp-ro,resolve))
$(eval $(call buildpack-mount,mod-verify))
$(eval $(call buildpack-mount,test))
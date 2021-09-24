APP_NAME = kyma-operator
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20210601-19326f24
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
override ENTRYPOINT := ./cmd/operator/main.go

include $(SCRIPTS_DIR)/generic-make-go.mk

verify:: vet mod-verify

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
	GO111MODULE=on go test -coverprofile=/tmp/artifacts/cover.out ./...
	@echo -n "Total coverage: "
	@go tool cover -func=/tmp/artifacts/cover.out | grep total | awk '{print $$3}'

$(eval $(call buildpack-mount,resolve))
$(eval $(call buildpack-mount,mod-verify))
$(eval $(call buildpack-mount,test))

.PHONY: clean
clean:
	rm -f $(binary)

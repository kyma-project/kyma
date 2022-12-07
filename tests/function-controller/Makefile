APP_NAME = function-controller-test
APP_PATH = tests/function-controller
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20220613-63e4233c
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

IMG_NAME := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG := $(DOCKER_TAG)

include $(SCRIPTS_DIR)/generic-make-go.mk

ENTRYPOINT := ./cmd/main.go
VERIFY_IGNORE := /vendor\|/mocks

test-local:
	GO111MODULE=on go test -count=1 -v ./pkg/...

resolve-local:
	go mod download
	go mod vendor

compile:
	CGO_ENABLED=0 go build -o main.test ./cmd/main.go

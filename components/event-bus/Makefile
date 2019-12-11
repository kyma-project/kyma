# Copyright 2019 The Kyma Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

APP_NAME = event-bus
APP_PATH = components/event-bus
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.12
SCRIPTS_DIR = ../../common/makefiles

# COMPONENT_DIR is a local path to component
include $(SCRIPTS_DIR)/generic-make-go.mk

# generate goal that runs 'all-local' inside a container
$(eval $(call buildpack-mount,all))

verify:: all

.DEFAULT_GOAL := all-local

# base import path
PKG := github.com/kyma-project/kyma/components/event-bus
# produce goals for each cmd program ("bin1 bin2 ...")
#cmds := $(foreach cmd,$(shell go list -f '{{ if eq .Name "main" }}{{ .ImportPath }}{{end}}' ./...),$(notdir $(cmd)))
cmds := $(foreach dir,$(wildcard cmd/*),$(notdir $(dir)))
# produce goals to build images for each cmd program ("bin1.image bin2.image ...")
cmds_images = $(foreach cmd,$(cmds),$(cmd).image)
# produce goals to push images for each cmd program ("bin1.image.push bin2.image.push ...")
cmds_images_push = $(foreach img,$(cmds_images),$(img).push)

# list all go imports that can be tested
# list all directories containing a Dockerfile
other_images := $(foreach img,$(wildcard ./images/*/Dockerfile),$(notdir $(patsubst %/,%,$(dir $(img)))))
# produce push targest for those
other_images_push = $(foreach img,$(other_images),$(img).push)

.PHONY: clean

all-local: clean $(cmds)

GOBUILD_FLAGS := -v
$(cmds): vendor
	@echo "+ Building $@"
	@CGO_ENABLED=0 go build -o $@ \
		$(GOBUILD_FLAGS) \
		$(PKG)/cmd/$@


clean:
	@echo "+ Cleaning"
	rm -rf bin/
	rm -f $(cmds)
	@go clean -x -i $(PKG)/cmd/...

.PHONY: test

vendor: Gopkg.lock
	@echo '+ Pulling vendored dependencies'
	@dep ensure -v --vendor-only


GOTEST_FLAGS := -cover
test-local: vendor
	@echo "+ Performing unit tests"
	$(eval testpkgs := $(shell go list -test -f '{{ if .TestGoFiles }}{{ .ForTest }}{{end}}' ./... ))
	@go test \
		$(GOTEST_FLAGS) \
		$(testpkgs)

.PHONY: $(other_images_push) $(other_images)
$(other_images_push): %.push: %
	$(eval $*_img_name := $(APP_NAME)-$*)
	$(eval IMG_NAME := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$($*_img_name))
	docker image tag $($*_img_name) $(IMG_NAME):$(TAG)
	docker image push $(IMG_NAME):$(TAG)

$(other_images): pull-licenses-local
	$(eval $@_img_name := $(APP_NAME)-$@)
	cp -R licenses ./images/$@
	@echo "+ Building container image $($@_img_name)"
	docker image build -t $($@_img_name) ./images/$@
	rm -rf  ./images/$@/licenses

.PHONY: $(cmds_images) $(cmds_images_push)

# override image goals from common Makefile because we need to build several images
build-image: ;
push-image: $(cmds_images_push) $(other_images_push);

# Example:
#   make controller-manager.image
$(cmds_images): %.image: % pull-licenses-local
	$(eval $*_img_name := $(APP_NAME)-$*)
	@echo "+ Building container image $($*_img_name)"
	docker image build -f cmd/$*/Dockerfile -t $($*_img_name) .

# Example:
#   make controller-manager.image.push
$(cmds_images_push): %.push: %
	$(eval $*_img_name := $(APP_NAME)-$(subst .image,,$*))
	@echo "+ Pushing container image $($*_img_name)"
	docker image tag $($*_img_name) $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$($*_img_name):$(TAG)
	docker image push $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$($*_img_name):$(TAG)

build-local: ;
ensure-local: ;

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/event-bus"

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

APP_NAME = end-to-end-backup
APP_PATH = tests/end-to-end/backup
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.13
SCRIPTS_DIR = ../../../common/makefiles

# COMPONENT_DIR is a local path to component
include $(SCRIPTS_DIR)/generic-make-go.mk

# generate goal that runs 'all-local' inside a container
$(eval $(call buildpack-mount,all))

verify:: all


#########
#       #
#  Dev  #
#       #
#########

.DEFAULT_GOAL := all-local

PKG := github.com/kyma-project/kyma/tests/end-to-end/backup

patched_trace := vendor/go.opencensus.io/trace/trace.go.patched

# produce goals for each cmd program ("bin1 bin2 ...")
cmds := $(foreach dir,$(wildcard test-cmd/*),$(notdir $(dir)))

bins := $(foreach cmd,$(cmds),$(cmd).test)
# produce goals to build images for each cmd program ("bin1.image bin2.image ...")
cmds_images = $(foreach cmd,$(cmds),$(cmd).image)
# produce goals to push images for each cmd program ("bin1.image.push bin2.image.push ...")
cmds_images_push = $(foreach img,$(cmds_images),$(img).push)


.PHONY: clean

all-local: clean $(cmds)

# Example:
#   make controller-manager
GOBUILD_FLAGS := -v
http-adapter: $(patched_trace)

$(cmds): vendor
	@echo "+ Building $@"
	@CGO_ENABLED=0 go test -o $@.test \
		$(GOBUILD_FLAGS) \
		-c $(PKG)/test-cmd/$@

clean:
	@echo "+ Cleaning"
	rm -rf bin/
	rm -f $(cmds)
	@go clean -x -i $(PKG)/test-cmd/...

vendor: Gopkg.lock
	@echo '+ Pulling vendored dependencies'
	@dep ensure -v --vendor-only


.PHONY: test

GOTEST_FLAGS := -cover
test-local: ;

.PHONY: $(cmds_images) $(cmds_images_push)

# override image goals from common Makefile because we need to build several images
build-image: ;
push-image: $(cmds_images_push) ;

# Example:
#   make controller-manager.image
$(cmds_images): %.image: %
	$(eval $@_img_name := $(APP_NAME)-$*-tests)
	@echo "+ Building container image $($@_img_name)"
	docker image build -f test-cmd/$*/Dockerfile -t $($@_img_name) .

# Example:
#   make controller-manager.image.push
$(cmds_images_push): %.push: %
	$(eval $@_img_name := $(APP_NAME)-$(subst .image,,$*)-tests)
	@echo "+ Pushing container image $($@_img_name)"
	docker image tag $($@_img_name) $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$($@_img_name):$(TAG)
	docker image push $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$($@_img_name):$(TAG)



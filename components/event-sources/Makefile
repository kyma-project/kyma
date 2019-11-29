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

APP_NAME = event-sources
APP_PATH = components/event-sources
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.12
SCRIPTS_DIR = ../../common/makefiles

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

PKG := github.com/kyma-project/kyma/components/event-sources

patched_trace := vendor/go.opencensus.io/trace/trace.go.patched

# produce goals for each cmd program ("bin1 bin2 ...")
cmds := $(foreach cmd,$(shell go list -f '{{ if eq .Name "main" }}{{ .ImportPath }}{{end}}' ./...),$(notdir $(cmd)))
# produce goals to build images for each cmd program ("bin1.image bin2.image ...")
cmds_images = $(foreach cmd,$(cmds),$(cmd).image)
# produce goals to push images for each cmd program ("bin1.image.push bin2.image.push ...")
cmds_images_push = $(foreach img,$(cmds_images),$(img).push)

# list all go imports that can be tested
testpkgs := $(shell go list -test -f '{{ if .TestGoFiles }}{{ .ForTest }}{{end}}' ./... )


.PHONY: clean

all-local: clean $(cmds)

# Example:
#   make controller-manager
GOBUILD_FLAGS := -v
http-adapter: $(patched_trace)
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

vendor: Gopkg.lock
	@echo '+ Pulling vendored dependencies'
	@dep ensure -v --vendor-only

# https://stackoverflow.com/a/13517811/4716370
$(patched_trace): hack/set-span-id.patch vendor
	@echo '+ Patching opencensus dependency'
	if [ -e $@ ]; then patch -R -p0 < $@; fi
	patch -p0 < $<
	cp $< $@


.PHONY: test


GOTEST_FLAGS := -cover
test-local: vendor $(patched_trace)
	@echo "+ Performing unit tests"
	@go test \
		$(GOTEST_FLAGS) \
		$(testpkgs)


.PHONY: $(cmds_images) $(cmds_images_push)

# override image goals from common Makefile because we need to build several images
build-image: ;
push-image: $(cmds_images_push) ;

# Example:
#   make controller-manager.image
$(cmds_images): %.image: %
	$(eval $@_img_name := $(APP_NAME)-$*)
	@echo "+ Building container image $($@_img_name)"
	docker image build -f cmd/$*/Dockerfile -t $($@_img_name) .

# Example:
#   make controller-manager.image.push
$(cmds_images_push): %.push: %
	$(eval $@_img_name := $(APP_NAME)-$(subst .image,,$*))
	@echo "+ Pushing container image $($@_img_name)"
	docker image tag $($@_img_name) $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$($@_img_name):$(TAG)
	docker image push $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$($@_img_name):$(TAG)


#####################
#                   #
#  Code generation  #
#                   #
#####################

# see https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md#generate-code

# e.g. "sources/v1alpha1 sources/v1alpha2"
API_GROUPS := sources/v1alpha1
# generates e.g. "PKG/apis/sources/v1alpha1 PKG/apis/sources/v1alpha2"
api-import-paths := $(foreach group,$(API_GROUPS),$(PKG)/apis/$(group))

generators := deepcopy client lister informer injection
generators_bin := $(foreach x,$(generators),bin/$(x)-gen)

.PHONY: codegen $(generators) $(generators_bin)

codegen: $(generators) injection

# http://blog.jgc.org/2007/06/escaping-comma-and-space-in-gnu-make.html
comma := ,
space :=
space +=

# doc: https://godoc.org/k8s.io/code-generator/cmd/deepcopy-gen
deepcopy: bin/deepcopy-gen
	@echo "+ Generating deepcopy funcs for $(API_GROUPS)"
	@bin/deepcopy-gen \
		--go-header-file hack/boilerplate.go.txt \
		--input-dirs $(subst $(space),$(comma),$(api-import-paths))

client: bin/client-gen
	@echo "+ Generating clientsets for $(API_GROUPS)"
	@rm -rf client/generated/clientset
	@bin/client-gen \
		--go-header-file hack/boilerplate.go.txt \
		--input $(subst $(space),$(comma),$(API_GROUPS)) \
		--input-base $(PKG)/apis \
		--clientset-path $(PKG)/client/generated/clientset

lister: bin/lister-gen
	@echo "+ Generating listers for $(API_GROUPS)"
	@rm -rf client/generated/lister
	@bin/lister-gen \
		--go-header-file hack/boilerplate.go.txt \
		--input-dirs $(subst $(space),$(comma),$(api-import-paths)) \
		--output-package $(PKG)/client/generated/lister

informer: bin/informer-gen
	@echo "+ Generating informers for $(API_GROUPS)"
	@rm -rf client/generated/informer
	@bin/informer-gen \
		--go-header-file hack/boilerplate.go.txt \
		--input-dirs $(subst $(space),$(comma),$(api-import-paths)) \
		--output-package $(PKG)/client/generated/informer \
		--versioned-clientset-package $(PKG)/client/generated/clientset/internalclientset \
		--listers-package $(PKG)/client/generated/lister

injection: bin/injection-gen
	@echo "+ Generating injection for $(API_GROUPS)"
	@rm -rf client/generated/injection
	@bin/injection-gen \
		--go-header-file hack/boilerplate.go.txt \
		--input-dirs $(subst $(space),$(comma),$(api-import-paths)) \
		--output-package $(PKG)/client/generated/injection \
		--versioned-clientset-package $(PKG)/client/generated/clientset/internalclientset \
		--external-versions-informers-package $(PKG)/client/generated/informer/externalversions

$(generators_bin): vendor
	@if [ -d ./vendor/k8s.io/code-generator/cmd/$(notdir $@) ]; then \
		go build -o bin/$(notdir $@) ./vendor/k8s.io/code-generator/cmd/$(notdir $@); \
	else \
		go build -o bin/$(notdir $@) ./vendor/knative.dev/pkg/codegen/cmd/$(notdir $@); \
	fi

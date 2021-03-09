APP_NAME = event-sources
APP_PATH = components/event-sources
SCRIPTS_DIR = ../../common/makefiles

# COMPONENT_DIR is a local path to component
include $(SCRIPTS_DIR)/generic-make-go.mk

# Override to use go modules
release:
	$(MAKE) gomod-release-local

# This target patch opencensus lib (add required method). This patch is required by knative.dev/eventing
gomod-deps-local:: vendor/go.opencensus.io/trace/trace.go.patched

#########
#       #
#  Dev  #
#       #
#########

PKG := github.com/kyma-project/kyma/components/event-sources

ISTIO_CLIENT_GO_PKG := istio.io/client-go/pkg

patched_trace := vendor/go.opencensus.io/trace/trace.go.patched

# produce goals for each cmd program ("bin1 bin2 ...")
cmds := $(foreach dir,$(wildcard cmd/*),$(notdir $(dir)))
# produce goals to build images for each cmd program ("bin1.image bin2.image ...")
cmds_images = $(foreach cmd,$(cmds),$(cmd).image)
cmds_clean = $(foreach cmd,$(cmds),$(cmd).clean)
# produce goals to push images for each cmd program ("bin1.image.push bin2.image.push ...")
cmds_images_push = $(foreach img,$(cmds_images),$(img).push)

#MOUNT_TARGETS = $(cmds)
#$(foreach t,$(MOUNT_TARGETS),$(eval $(call buildpack-mount,$(t))))

.PHONY: clean
clean: $(cmds_clean) resolve_clean licenses_clean

$(cmds_clean): %.clean:
	rm -f $*

resolve_clean:
	rm -rf vendor

licenses_clean:
	rm -rf licenses

build-local: $(cmds) test-local ;

# Example:
#   make controller-manager
GOBUILD_FLAGS := -v
$(cmds): %: gomod-deps-local $(patched_trace)
	@echo "+ Building $*"
	@CGO_ENABLED=0 go build -o $* \
		$(GOBUILD_FLAGS) \
		$(PKG)/cmd/$*

# https://stackoverflow.com/a/13517811/4716370
$(patched_trace): hack/set-span-id.patch vendor
	@echo '+ Patching opencensus dependency'
	if [ -e $@ ]; then patch -R -p0 < $@; fi
	patch -p0 < $<
	cp $< $@

test-local: gomod-vendor-local $(patched_trace)

.PHONY: $(cmds_images) $(cmds_images_push)

# override image goals from common Makefile because we need to build several images
build-image: gomod-deps $(patched_trace) $(cmds_images) ;
push-image: $(cmds_images_push) ;

# Example:
#   make controller-manager.image
$(cmds_images): %.image: pull-licenses
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

# e.g. "authentication/v1alpha1 networking/v1alpha3"
ISTIO_API_GROUPS := "security/v1beta1"
# generates e.g. "ISTIO_CLIENT_GO_PKG/apis/sources/v1alpha1 ISTIO_CLIENT_GO_PKG/apis/sources/v1alpha2"
istio-api-import-paths := $(foreach group,$(ISTIO_API_GROUPS),$(ISTIO_CLIENT_GO_PKG)/apis/$(group))

generators := deepcopy client lister informer injection istio-injection
generators_bin := $(foreach x,$(generators),bin/$(x)-gen)

.PHONY: codegen $(generators) $(generators_bin)

codegen: $(generators)

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

istio-injection: bin/injection-gen
	@echo "+ Generating injection for $(ISTIO_API_GROUPS)"
	@rm -rf client/generated/injection/istio
	@bin/injection-gen \
        --go-header-file hack/boilerplate.go.txt \
        --input-dirs $(subst $(space),$(comma),$(istio-api-import-paths)) \
        --output-package $(PKG)/client/generated/injection/istio \
        --versioned-clientset-package $(ISTIO_CLIENT_GO_PKG)/clientset/versioned \
        --external-versions-informers-package $(ISTIO_CLIENT_GO_PKG)/informers/externalversions

$(generators_bin): gomod-vendor-local
	@if [ -d ./vendor/k8s.io/code-generator/cmd/$(notdir $@) ]; then \
		go build -o bin/$(notdir $@) ./vendor/k8s.io/code-generator/cmd/$(notdir $@); \
	else \
		go build -o bin/$(notdir $@) ./vendor/knative.dev/pkg/codegen/cmd/$(notdir $@); \
	fi

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/event-sources"

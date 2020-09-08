APP_NAME = application-broker
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles
PKG := github.com/kyma-project/kyma/components/application-broker

include $(SCRIPTS_DIR)/generic-make-go.mk

#ENTRYPOINT:=cmd/broker/main.go
VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg

# produce goals for each cmd program ("bin1 bin2 ...")
cmds := $(foreach dir,$(wildcard cmd/*),$(notdir $(dir)))
# produce goals to build images for each cmd program ("bin1.image bin2.image ...")
cmds_images = $(foreach cmd,$(cmds),$(cmd).image)
cmds_clean = $(foreach cmd,$(cmds),$(cmd).clean)
# produce goals to push images for each cmd program ("bin1.image.push bin2.image.push ...")
cmds_images_push = $(foreach img,$(cmds_images),$(img).push)

MOUNT_TARGETS = $(cmds)
$(foreach t,$(MOUNT_TARGETS),$(eval $(call buildpack-mount,$(t))))

build-local: $(cmds) test-local ;

GOBUILD_FLAGS := -v
$(cmds): %: resolve-local
	@echo "+ Building $*"
	@CGO_ENABLED=0 go build -o $* \
		$(GOBUILD_FLAGS) \
		$(PKG)/cmd/$*

.PHONY: clean
clean: $(cmds_clean) resolve_clean licenses_clean

$(cmds_clean): %.clean:
	rm -f $*

resolve_clean:
	rm -rf vendor

licenses_clean:
	rm -rf licenses

# Example:
#   make controller-manager.image
$(cmds_images): %.image: build pull-licenses
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

# override image goals from common Makefile because we need to build several images
build-image: $(cmds_images) ;
push-image: $(cmds_images_push) ;

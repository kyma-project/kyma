BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.12
SCRIPTS_DIR = ../../common/makefiles
APP_PATH = tests/event-bus
APP_NAME = event-bus
# COMPONENT_DIR is a local path to component
include $(SCRIPTS_DIR)/generic-make-go.mk

images = e2e-subscriber e2e-tester
img_image := $(foreach img,$(images),$(img).image)
img_push := $(foreach img,$(images),$(img).push)
$(foreach img,$(images),$(eval $(img).image_name := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)-$(img)))

build-image: $(img_build);
push-image: $(img_push);

test: ;

$(img_push): %.push: %.image
	docker tag $($*.image_name) $($*.image_name):$(TAG)
	docker push  $($*.image_name):$(TAG)

$(img_image): %.image: pull-licenses
	docker build -t $($*.image_name) -f $*/Dockerfile .

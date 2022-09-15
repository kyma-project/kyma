APP_NAME = event-publisher-proxy
APP_PATH = components/$(APP_NAME)
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20220407-4da6c929
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

# fail on lint issues
override IGNORE_LINTING_ISSUES =
override ENTRYPOINT = cmd/main.go

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

release:
	$(MAKE) gomod-release-local

path-to-referenced-charts:
	@echo "resources/event-publisher-proxy"

# produce goals for each cmd program ("bin1 bin2 ...")
cmds := $(foreach dir,$(wildcard cmd/*),$(notdir $(dir)))
# produce goals to build images for each cmd program ("bin1.image bin2.image ...")
cmds_images = $(foreach cmd,$(cmds),$(cmd).image)
cmds_clean = $(foreach cmd,$(cmds),$(cmd).clean)
# produce goals to push images for each cmd program ("bin1.image.push bin2.image.push ...")
cmds_images_push = $(foreach img,$(cmds_images),$(img).push)

.PHONY: clean
clean: $(cmds_clean) resolve_clean

$(cmds_clean): %.clean:
	rm -f $*

resolve_clean:
	rm -rf vendor

build-local: $(cmds) test-local ;

.PHONY: $(cmds_images) $(cmds_images_push)

# override image goals from common Makefile because we need to build several images
build-image: $(cmds_images) ;
push-image: $(cmds_images_push) ;

# Example:
#   make event-publisher-proxy
GOBUILD_FLAGS := -v
PKG := github.com/kyma-project/kyma/components/event-publisher-proxy
$(cmds): %:
	@echo "+ Building $*"
	@CGO_ENABLED=0 go build -o $* \
		$(GOBUILD_FLAGS) \
		$(PKG)/cmd/$*

# Example:
#   make event-publisher-proxy.image
$(cmds_images): %.image: build-local
	$(eval $@_img_name := $*)
	@echo "+ Building container image $($@_img_name)"
	docker image build -f cmd/$*/Dockerfile -t $($@_img_name) .

# Example:
#   make event-publisher-proxy.image.push
$(cmds_images_push): %.push: %
	$(eval $@_img_name := $(subst .image,,$*))
	@echo "+ Pushing container image $($@_img_name)"
	docker image tag $($@_img_name) $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$($@_img_name):$(TAG)
	docker image push $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$($@_img_name):$(TAG)

## Run all tests
test-local:
	mkdir -p /tmp/artifacts
	go test -coverprofile=/tmp/artifacts/cover.out --tags="unit integration" ./...
	@echo -n "Total coverage: "
	@go tool cover -func=/tmp/artifacts/cover.out | grep total | awk '{print $$3}'

## Run only unit tests
test-local-unit:
	mkdir -p /tmp/artifacts
	go test -coverprofile=/tmp/artifacts/cover.out --tags=unit ./...
	@echo -n "Total coverage: "
	@go tool cover -func=/tmp/artifacts/cover.out | grep total | awk '{print $$3}'

## Run only integration tests
test-local-integration:
	mkdir -p /tmp/artifacts
	go test -coverprofile=/tmp/artifacts/cover.out --tags=integration ./...
	@echo -n "Total coverage: "
	@go tool cover -func=/tmp/artifacts/cover.out | grep total | awk '{print $$3}'

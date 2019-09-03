IMG_NAME = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)
BASE_PKG = github.com/kyma-project/kyma
IMG_GOPATH = /workspace/go
IMG_GOCACHE = /root/.cache/go-build
VERIFY_IGNORE = /vendor\|/automock\|/testdata\|/pkg

LOCAL_DIR = $(shell pwd)
WORKSPACE_DIR = $(IMG_GOPATH)/src/$(BASE_PKG)/$(APP_PATH)
FILES_TO_CHECK = find . -type f -name "*.go" | grep -v "$(VERIFY_IGNORE)"
DIRS_TO_CHECK = go list ./... | grep -v "$(VERIFY_IGNORE)"
DIRS_TO_IGNORE = go list ./... | grep "$(VERIFY_IGNORE)"

.PHONY: release resolve build build-image push-image pull-licenses

DOCKER_CREATE_OPTS := --rm -w $(WORKSPACE_DIR) $(BUILDPACK)
ifndef GOPATH
DOCKER_CREATE_OPTS := -v $(shell go env GOCACHE):$(IMG_GOCACHE):delegated -v $(shell go env GOPATH)/pkg/dep:$(IMG_GOPATH)/pkg/dep:delegated -t $(DOCKER_CREATE_OPTS)
DOCKER_INTERACTIVE := -i
endif

define buildpack-mount
	@docker run $(DOCKER_INTERACTIVE) -v $(LOCAL_DIR):$(WORKSPACE_DIR):delegated $(DOCKER_CREATE_OPTS) make $(1)-local
endef

define buildpack-cp-ro
	$(eval container = $(shell docker create $(DOCKER_CREATE_OPTS) make $(1)-local))
	@docker cp $(LOCAL_DIR)/. $(container):$(WORKSPACE_DIR)/
	@docker start $(DOCKER_INTERACTIVE) $(container)
endef

verify: test imports fmt
verify-local: test-local imports-local fmt-local dep-status-local
release: resolve dep-status verify build-image push-image

build-image: pull-licenses
	docker build -t $(IMG_NAME) .

push-image:
	docker tag $(IMG_NAME) $(IMG_NAME):$(TAG)
	docker push $(IMG_NAME):$(TAG)

build:
	$(call buildpack-mount,build)
build-local:
	env CGO_ENABLED=0 go build -o $(APP_NAME)
	rm $(APP_NAME)

resolve:
	$(call buildpack-mount,resolve)
resolve-local:
	dep ensure -vendor-only

ensure:
	$(call buildpack-mount,ensure)
ensure-local:
	dep ensure

dep-status:
	$(call buildpack-mount,dep-status)
dep-status-local:
	dep status

test:
	$(call buildpack-cp-ro,test)
test-local:
	go test ./...

imports:
	$(call buildpack-mount,imports)
imports-local:
	goimports -w -l $$($(FILES_TO_CHECK))

fmt:
	$(call buildpack-mount,fmt)
fmt-local:
	go fmt $$($(DIRS_TO_CHECK))

errcheck:
	$(call buildpack-mount,errcheck)
errcheck-local:
	errcheck -blank -asserts -ignorepkg '$$($(DIRS_TO_IGNORE) | tr '\n' ',')' -ignoregenerated ./...

vet:
	$(call buildpack-mount,vet)
vet-local:
	go vet $$($(DIRS_TO_CHECK))

pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif
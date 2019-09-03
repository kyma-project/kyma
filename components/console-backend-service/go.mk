# Default configuration
IMG_NAME = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)
BASE_PKG = github.com/kyma-project/kyma
IMG_GOPATH = /workspace/go
IMG_GOCACHE = /root/.cache/go-build
VERIFY_IGNORE = /vendor\|/automock\|/testdata\|/pkg

# Other variables
LOCAL_DIR = $(shell pwd)
WORKSPACE_DIR = $(IMG_GOPATH)/src/$(BASE_PKG)/$(APP_PATH)
FILES_TO_CHECK = find . -type f -name "*.go" | grep -v "$(VERIFY_IGNORE)"
DIRS_TO_CHECK = go list ./... | grep -v "$(VERIFY_IGNORE)"
DIRS_TO_IGNORE = go list ./... | grep "$(VERIFY_IGNORE)"
INTERACTIVE := $(shell [ -t 0 ] && echo 1)

# Docekr configuration
DOCKER_CREATE_OPTS := --rm -w $(WORKSPACE_DIR) $(BUILDPACK)

ifdef GOPATH
DOCKER_CREATE_OPTS := -v $(shell go env GOCACHE):$(IMG_GOCACHE):delegated -v $(shell go env GOPATH)/pkg/dep:$(IMG_GOPATH)/pkg/dep:delegated $(DOCKER_CREATE_OPTS)
endif

ifdef INTERACTIVE
DOCKER_INTERACTIVE := -i
DOCKER_CREATE_OPTS := -t $(DOCKER_CREATE_OPTS)
endif

# Buildpack directives
define buildpack-mount
	@echo make $(1)
	@docker run $(DOCKER_INTERACTIVE) -v $(LOCAL_DIR):$(WORKSPACE_DIR):delegated $(DOCKER_CREATE_OPTS) make $(1)-local
endef

define buildpack-cp-ro
	$(eval container = $(shell docker create $(DOCKER_CREATE_OPTS) make $(1)-local))
	@docker cp $(LOCAL_DIR)/. $(container):$(WORKSPACE_DIR)/
	@docker start $(DOCKER_INTERACTIVE) $(container)
endef

.PHONY: verify format release
verify: test check-imports check-fmt
format: imports fmt
release: resolve dep-status verify build-image push-image

.PHONY: build-image push-image
build-image: pull-licenses
	docker build -t $(IMG_NAME) .
push-image:
	docker tag $(IMG_NAME) $(IMG_NAME):$(TAG)
	docker push $(IMG_NAME):$(TAG)

.PHONY: build build-local
build:
	$(call buildpack-mount,build)
build-local:
	env CGO_ENABLED=0 go build -o $(APP_NAME)
	rm $(APP_NAME)

.PHONY: resolve resolve-local
resolve:
	$(call buildpack-mount,resolve)
resolve-local:
	dep ensure -vendor-only

.PHONY: ensure ensure-local
ensure:
	$(call buildpack-mount,ensure)
ensure-local:
	dep ensure

.PHONY: dep-status dep-status-local
dep-status:
	$(call buildpack-mount,dep-status)
dep-status-local:
	dep status

.PHONY: test test-local
test:
	$(call buildpack-cp-ro,test)
test-local:
	go test ./...

.PHONY: check-imports check-imports-local
check-imports:
	$(call buildpack-mount,check-imports)
check-imports-local:
	exit $(shell goimports -l $$($(FILES_TO_CHECK)) | wc -l | xargs)

.PHONY: imports imports-local
imports:
	$(call buildpack-mount,imports)
imports-local:
	goimports -w -l $$($(FILES_TO_CHECK))

.PHONY: check-fmt check-fmt-local
check-fmt:
	$(call buildpack-mount,check-fmt)
check-fmt-local:
	exit $(shell gofmt -l $$($(FILES_TO_CHECK)) | wc -l | xargs)

.PHONY: fmt fmt-local
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
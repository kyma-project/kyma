IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)
BASE_PKG = github.com/kyma-project/kyma
IMG_GOPATH = /workspace/go
IMG_GOCACHE = /root/.cache/go-build
PKGS_TO_IGNORE = /vendor\|/automock\|/testdata\|/pkg

LOCAL_DIR = $(shell pwd)
WORKSPACE_DIR = $(IMG_GOPATH)/src/$(BASE_PKG)/$(APP_PATH)
FILES_TO_CHECK = $(find . -type f -name "*.go" | grep -v "$(PKGS_TO_IGNORE)"')
PKGS_TO_CHECK = $(go list ./... | grep -v "$(PKGS_TO_IGNORE)")

.PHONY: release resolve build build-image push-image pull-licenses

DOCKER_CREATE_OPTS = --rm -t -w $(WORKSPACE_DIR) -v $(shell go env GOCACHE):$(IMG_GOCACHE) -v $(shell go env GOPATH)/pkg/dep:$(IMG_GOPATH)/pkg/dep $(BUILDPACK)

define buildpack-mount
	@docker run -i -v $(LOCAL_DIR):$(WORKSPACE_DIR) $(DOCKER_CREATE_OPTS) make $(1)-local
endef

define buildpack-cp
	$(eval container = $(shell docker create $(DOCKER_CREATE_OPTS) make $(1)-local))
	@docker cp $(LOCAL_DIR)/. $(container):$(WORKSPACE_DIR)/
	@docker start -i $(container)
endef

verify: dep-status test imports fmt errcheck vet
release: resolve verify build-image push-image

build-image: pull-licenses
	docker build -t $(IMG_NAME) .

push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

build:
	$(call buildpack-cp,build)
build-local:
	env CGO_ENABLED=0 go build -o $(APP_NAME)
	rm $(APP_NAME)

resolve:
	$(call buildpack-cp,resolve)
resolve-local:
	dep ensure -v -vendor-only

dep-status:
	$(call buildpack-cp,dep-status)
dep-status-local:
	dep status -v

test:
	$(call buildpack-cp,test)
test-local:
	go test ./...

imports:
	$(call buildpack-mount,imports)
imports-local:
	goimports -w -l $(FILES_TO_CHECK))

fmt:
	$(call buildpack-mount,go fmt $(PKGS_TO_CHECK))
fmt-local:
	go fmt $(PKGS_TO_CHECK)

errcheck:
	$(call buildpack-mount,errcheck)
errcheck-local:
	errcheck -blank -asserts -ignoregenerated -exclude automock ./...

vet:
	$(call buildpack-mount,vet)
vet-local:
	go vet $(PKGS_TO_CHECK)

pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif
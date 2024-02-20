APP_NAME = common
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: build
build:
	go mod vendor
	go mod verify
	go test -count 100 -race -coverprofile=cover.out ./...

testi:
.PHONY: test
test: build

.PHONY: ci-pr
ci-pr: build

.PHONY: ci-main
ci-main: build

.PHONY: ci-release
ci-release: build

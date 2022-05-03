REPO_VERSION=$(shell git rev-list -1 HEAD)

.PHONY: build-image
build-image:
	docker build . -t kyma-eventing-fit:$(REPO_VERSION)
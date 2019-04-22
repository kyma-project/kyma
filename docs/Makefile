IMG_DIR = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/
TAG = $(DOCKER_TAG)
DOCS_BUILDER_PATH=../tools/docsbuilder

ci-pr: resolve-tool build-and-push-images
ci-master: resolve-tool build-and-push-images prepare-content
ci-release: resolve-tool build-and-push-images prepare-content

resolve-tool:
	cd ${DOCS_BUILDER_PATH} && \
	dep ensure -v -vendor-only

build-and-push-images:
	cd ${DOCS_BUILDER_PATH} && \
	APP_DOCKER_IMAGE_PREFIX=${IMG_DIR} \
	APP_DOCKER_IMAGE_TAG=${TAG} \
	go run main.go

prepare-content:
	make -C "../../website" prepare-content

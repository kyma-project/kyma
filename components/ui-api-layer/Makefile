IMG_NAME = ui-api-layer
IMG = $(DOCKER_PUSH_REPOSITORY)/$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

resolve:
	dep ensure -v -vendor-only

build:
	./before-commit.sh ci

unit-tests:
	go test ./...

build-image:
	docker build -t $(IMG_NAME) .

push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)
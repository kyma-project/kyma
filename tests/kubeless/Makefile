APP_NAME = kubeless-tests
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)


resolve: 
	dep ensure -vendor-only -v

validate: resolve
	go build -o golint-vendored ./vendor/github.com/golang/lint/golint
	./golint-vendored
	rm golint-vendored

build: resolve
	docker build -t $(APP_NAME):latest .

push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

.PHONY: ci-pr
ci-pr: resolve validate build push-image

.PHONY: ci-master
ci-master: resolve validate build push-image

.PHONY: ci-release
ci-release: resolve validate build push-image

#.PHONY: path-to-referenced-charts
#path-to-referenced-charts:
#	@echo "resources/core"

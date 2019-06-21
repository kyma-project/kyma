APP_NAME = binding-usage-controller
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)
binary=$(APP_NAME)

APP_MIGRATION = service-catalog-upgrade
IMG_MIGRATION = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_MIGRATION)
binary_migration=$(APP_MIGRATION)

.PHONY: build
build:
	./before-commit.sh ci

.PHONY: build-image
build-image:
	cp $(binary) deploy/controller/$(binary)
	docker build -t $(APP_NAME):latest deploy/controller

	cp $(binary_migration) deploy/migration/$(binary_migration)
	docker build -t $(APP_MIGRATION):latest deploy/migration

.PHONY: push-image
push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

	docker tag $(APP_MIGRATION) $(IMG_MIGRATION):$(TAG)
	docker push $(IMG_MIGRATION):$(TAG)

.PHONY: ci-pr
ci-pr: build build-image push-image

.PHONY: ci-master
ci-master: build build-image push-image

.PHONY: ci-release
ci-release: build build-image push-image

.PHONY: clean
clean:
	rm -f $(binary)


.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/service-catalog-addons"

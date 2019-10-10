APP_NAME = application-broker
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)
binary=$(APP_NAME)

APP_MIGRATION = ab-upgrade-protector
IMG_MIGRATION = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_MIGRATION)
binary_migration=$(APP_MIGRATION)


.PHONY: build
build:
	./before-commit.sh ci

.PHONY: pull-licenses
pull-licenses:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

.PHONY: build-image
build-image: pull-licenses
	cp $(binary) deploy/broker/$(binary)
	cp $(binary_migration) deploy/migration/$(binary_migration)
	cp -R licenses deploy/broker/

	docker build -t $(APP_NAME):latest deploy/broker
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
	@echo "resources/application-connector"


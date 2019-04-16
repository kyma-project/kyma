APP_NAME = backup-restore-test
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/tests/$(APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: build
build:
	./before-commit.sh

.PHONY: build-image
build-image:
	mkdir -p build-temp
	cp ../../../docs/backup/docs/assets/all-backup.yaml build-temp
	cp ../../../docs/backup/docs/assets/system-backup.yaml build-temp
	docker build -t $(APP_NAME):latest .
	rm -rf build-temp

.PHONY: push-image
push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

.PHONY: ci-pr
ci-pr: build build-image push-image

.PHONY: ci-master
ci-master: build build-image push-image

.PHONY: ci-release
ci-release: build build-image push-image
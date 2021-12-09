APP_NAME = dex-static-user-configurer
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: build-image
build-image:
	docker build -t $(APP_NAME):latest .

.PHONY: push-image
push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)
ifeq ($(JOB_TYPE), postsubmit)
	@echo "Sign image with Cosign"
	cosign version
	cosign sign -key ${KMS_KEY_URL} $(IMG):$(TAG)
else
	@echo "Image signing skipped"
endif

release: build-image push-image component_descriptor_test

component_descriptor_test:
	gcloud auth configure-docker europe-west3-docker.pkg.dev > /dev/null 2>&1

	git clone https://github.com/Halamix2/test-infra.git ti
	mkdir -p /logs/artifacts/componentdescriptor

	cd ti; git checkout component_descriptor_generator; \
	go run development/tools/cmd/componentdescriptorgenerator/main.go \
	--component-name "github.com/kyma-project/kyma/$(APP_NAME)" \
	--component-version "0.0.1-dev-$(PULL_PULL_SHA)" \
	--app-name $(APP_NAME) \
	--output-dir /logs/artifacts/componentdescriptor \
	--image-url "$(IMG):$(TAG)" \
	--repo-context "$(DOCKER_PUSH_REPOSITORY)"

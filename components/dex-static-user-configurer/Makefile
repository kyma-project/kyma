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
	echo "name: kyma-project_kyma_components_yaml\nversion: \"0.0.1\"\ntype: \"ociImage\"\nrelation: \"local\"\naccess:\n  type: \"ociRegistry\"\n  url: \"$(IMG):$(TAG)\"\n---" >>  oci-image.yaml
	git clone https://github.com/Halamix2/test-infra.git ti
	cd ti
	git remote add origin https://github.com/Halamix2/test-infra.git
	git fetch origin > /dev/null
	git checkout component_descriptor_generator
	go run development/tools/cmd/componentdescriptorgenerator/main.go \
	--component-name "github.com/kyma-project/kyma/dex-static-configurer-tests" \
    --component-version "0.0.1" \
    --output-dir /tmp/artifacts \
	--resource oci-image.yaml \
    # --sources git.yaml \
    # --component-references component-reference.yaml \

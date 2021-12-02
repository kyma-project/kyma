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
	echo "---
name: kyma-project_kyma_components_yaml
version: \"0.0.1\"
type: \"ociImage\"
relation: \"local\"
access:
  type: \"ociRegistry\"
  url: \"$(IMG):$(TAG)\"
---" >>  oci-image.yaml
	git clone git@github.com:Halamix2/test-infra.git
	cd test-infra
	git checkout component_descriptor_generator
	go run development/tools/cmd/componentdescriptorgenerator/ \
	--component-name "github.com/kyma-project/kyma/dex-static-configurer-tests" \
    --component-version "0.0.1" \
    --output-dir test_output \
	--resource oci-image.yaml \
    # --sources git.yaml \
    # --component-references component-reference.yaml \

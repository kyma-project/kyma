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
	$(eval DOCKER_SHA := $(docker inspect --format='{{index .RepoDigests 0}}' $(IMG):$(TAG)))
	echo "name: kyma-project_kyma_components_$(APP_NAME)\nversion: \"0.0.1\"\ntype: \"ociImage\"\nrelation: \"local\"\naccess:\n  type: \"ociRegistry\"\n  url: \"$(IMG):$(DOCKER_SHA)\"\n---" >>  oci-image.yaml
	echo "---\nname: kyma-project_kyma_$(APP_NAME)\nversion: 0.0.1\ntype: git\naccess:\n  type: github\n  repoUrl: https://github.com/kyma-project/kyma/$(APP_NAME)\n  ref: main\n  commit: $(PULL_PULL_SHA)" >>  git.yaml
	git clone https://github.com/Halamix2/test-infra.git ti
	mkdir -p /tmp/artifacts/componentdescriptor

	cd ti; git checkout component_descriptor_generator; \
	go run development/tools/cmd/componentdescriptorgenerator/main.go \
	--component-name "github.com/kyma-project/kyma/$(APP_NAME)" \
	--component-version "0.0.1" \
	--output-dir /tmp/artifacts/componentdescriptor \
	--resource ../oci-image.yaml \
	--sources ../git.yaml

	cat /tmp/artifacts/componentdescriptor/component-descriptor.yaml

	curl -L https://github.com/gardener/component-cli/releases/download/v0.31.0/componentcli-linux-amd64.gz | gzip -d > ./component-cli; chmod +x component-cli
	gcloud auth configure-docker europe-west3-docker.pkg.dev > /dev/null 2>&1
	./component-cli ca remote push /tmp/artifacts/componentdescriptor --repo-ctx "europe-west3-docker.pkg.dev/sap-kyma-prow-workloads/test-gcr"

APP_NAME = compass-runtime-agent
APP_PATH = components/$(APP_NAME)
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/common/makefiles

override ENTRYPOINT = cmd/

include $(SCRIPTS_DIR)/generic-make-go.mk

VERIFY_IGNORE := /vendor\|/mocks

resolve-local:
	GO111MODULE=on go mod vendor -v

release:
	$(MAKE) gomod-release-local

.PHONY: path-to-referenced-charts
path-to-referenced-charts:
	@echo "resources/compass-runtime-agent"

.PHONY: test
test: envtest ## Run tests.
	if [ -d "$(ARTIFACTS)" ]; then \
		KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -v -coverprofile $(ARTIFACTS)/filtered.cov; \
	else \
		KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -v; \
	fi


ENVTEST_K8S_VERSION = 1.26.0

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

ENVTEST ?= $(LOCALBIN)/setup-envtest

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

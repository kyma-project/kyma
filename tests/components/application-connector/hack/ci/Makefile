K3D_URL=https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh

PROJECT_ROOT  ?= ../../../../..
CLUSTER_NAME  ?= kyma
REGISTRY_PORT ?= 5001
REGISTRY_NAME ?= ${CLUSTER_NAME}-registry

# Operating system architecture
OS_ARCH ?= $(shell uname -m)

# Operating system type
OS_TYPE ?= $(shell uname)

.PHONY: setup-environment run-gateway-tests run-validator-tests run-agent-test install-k3d 

.PHONY: install-k3d
install-k3d:
	curl --silent --fail ${K3D_URL} | TAG=${DEFAULT_K3D_VERSION} bash

setup-environment: kyma install-k3d ## Setup environment for tests
	${KYMA} provision k3d \
		--registry-port ${REGISTRY_PORT} \
		--name ${CLUSTER_NAME} \
		-p 8080:80@loadbalancer \
		-p 8443:443@loadbalancer \
		--ci

k3d-gateway-tests: setup-environment
	${KYMA} deploy \
		--ci --components-file ../../resources/installation-config/mini-kyma-os.yaml \
		--source=local \
		--workspace ${PROJECT_ROOT}
	make -f Makefile.test-application-gateway test
	k3d cluster delete

k3d-validator-tests: setup-environment
	${KYMA} deploy --ci --components-file ${PROJECT_ROOT}/resources/installation-config/mini-kyma-skr.yaml --value global.disableLegacyConnectivity=true --source=local --workspace ${KYMA_ROOT_CI}
	cd ${PROJECT_ROOT}
	make -f Makefile.test-application-conn-validator test
	k3d cluster delete

k3d-agent-tests: setup-environment
	${KYMA} deploy --ci --components-file ${PROJECT_ROOT}/resources/installation-config/mini-kyma-skr.yaml --value global.disableLegacyConnectivity=true --value compassRuntimeAgent.director.proxy.insecureSkipVerify=true --source=local --workspace ${KYMA_ROOT_CI}
	kubectl apply -f ${PROJECT_ROOT}/resources/patches/coredns.yaml
	kubectl -n kube-system delete pods -l k8s-app=kube-dns
	cd ${PROJECT_ROOT}
	make -f Makefile.test-compass-runtime-agent test
	k3d cluster delete

##@ Tools

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

########## Kyma CLI ###########
KYMA_STABILITY ?= unstable

define os_error
$(error Error: unsuported platform OS_TYPE:$1, OS_ARCH:$2; to mitigate this problem set variable KYMA with absolute path to kyma-cli binary compatible with your operating system and architecture)
endef

KYMA_FILE_NAME ?= kyma-linux

KYMA ?= $(LOCALBIN)/kyma-$(KYMA_STABILITY)
kyma: $(LOCALBIN) $(KYMA) ## Download kyma locally if necessary.
$(KYMA):
	$(if $(KYMA_FILE_NAME),,$(call os_error, ${OS_TYPE}, ${OS_ARCH}))
	test -f $@ || curl -s -Lo $(KYMA) https://storage.googleapis.com/kyma-cli-$(KYMA_STABILITY)/$(KYMA_FILE_NAME)
	chmod 0100 $(KYMA)


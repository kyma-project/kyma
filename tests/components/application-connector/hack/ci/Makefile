PROJECT_ROOT ?= ../..
KYMA_ROOT_CI ?= /home/prow/go/src/github.com/kyma-project/kyma
.PHONY: setup-environment run-gateway-tests run-validator-tests run-agent-test

.ONESHELL:
setup-environment:
	set -e
	k3d registry create k3d-registry --port 5000
	k3d cluster create k3d --registry-use k3d-k3d-registry:5000
	kubectl cluster-info
	CLI_VERSION=$(shell curl -s https://api.github.com/repos/kyma-project/cli/releases/latest | grep tag_name | cut -d '"' -f 4); \
	curl -Lo kyma.tar.gz https://github.com/kyma-project/cli/releases/download/$$CLI_VERSION/kyma_Linux_x86_64.tar.gz && mkdir kyma-release && tar -C kyma-release -zxvf kyma.tar.gz && chmod +x kyma-release/kyma && mv kyma-release/kyma /usr/local/bin && rm -rf kyma-release kyma.tar.gz

k3d-gateway-tests: setup-environment
	apk add openssl
	kyma deploy --ci --components-file ${PROJECT_ROOT}/resources/installation-config/mini-kyma-os.yaml --source=local --workspace ${KYMA_ROOT_CI}
	cd ${PROJECT_ROOT}
	make -f Makefile.test-application-gateway test
	k3d cluster delete


k3d-validator-tests: setup-environment
	kyma deploy --ci --components-file ${PROJECT_ROOT}/resources/installation-config/mini-kyma-skr.yaml --value global.disableLegacyConnectivity=true --source=local --workspace ${KYMA_ROOT_CI}
	cd ${PROJECT_ROOT}
	make -f Makefile.test-application-conn-validator test
	k3d cluster delete

k3d-agent-tests: setup-environment
	kyma deploy --ci --components-file ${PROJECT_ROOT}/resources/installation-config/mini-kyma-skr.yaml --value global.disableLegacyConnectivity=true --value compassRuntimeAgent.director.proxy.insecureSkipVerify=true --source=local --workspace ${KYMA_ROOT_CI}
	kubectl apply -f ${PROJECT_ROOT}/resources/patches/coredns.yaml
	kubectl -n kube-system delete pods -l k8s-app=kube-dns
	cd ${PROJECT_ROOT}
	make -f Makefile.test-compass-runtime-agent test
	k3d cluster delete
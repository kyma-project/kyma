NAMESPACE ?= test
GOPATH ?= $(shell go env GOPATH)

GATEWAY_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(DOCKER_TAG)"
VALIDATOR_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/connectivity-validator-test:$(DOCKER_TAG)"
COMPASS_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/compass-runtime-agent-test:$(DOCKER_TAG)"
MOCK_APP_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/mock-app:$(DOCKER_TAG)"
MOCK_SERVICE_NAME="mock-application"
APP_URL = "$(MOCK_SERVICE_NAME).$(NAMESPACE).svc.cluster.local"
TEST_TIMEOUT = "3m"

.PHONY: release test image clean
.PHONY: test-gateway test-gateway-debug clean-gateway-test publish-gateway-test image-gateway-test publish-mock-app image-mock-app disable-sidecar-for-mtls-test enable-sidecar-after-mtls-test generate-certs
.PHONY: test-compass-runtime-agent test-compass-runtime-agent-debug clean-compass-runtime-agent-test image-compass-runtime-agent-test publish-compass-runtime-agent-test
.PHONY: patch-for-validator-test unpatch-after-validator-test test-validator test-validator-debug validator-create-resources clean-validator-test publish-validator-test

release: publish-gateway-test publish-mock-app publish-validator-test publish-compass-runtime-agent-test
test: test-gateway test-validator test-compass-runtime-agent
image: image-gateway-test image-validator-test image-compass-runtime-agent-test
clean: clean-gateway-test clean-validator-test clean-compass-runtime-agent-test

test-gateway: disable-sidecar-for-mtls-test generate-certs gateway-create-resources
	if kubectl wait --for=condition=complete --timeout=$(TEST_TIMEOUT) -n $(NAMESPACE) job/application-gateway-test; then \
	echo "Success! Results:"; \
	./scripts/check-pod-logs.sh application-gateway-test; \
	$(MAKE) clean-gateway-test; \
	else \
	echo "Tests failed! Results:"; \
	./scripts/check-pod-logs.sh application-gateway-test; \
	$(MAKE) clean-gateway-test; \
	exit 1; \
	fi

test-gateway-debug: disable-sidecar-for-mtls-test generate-certs gateway-create-resources
	kubectl wait --for=condition=complete --timeout=$(TEST_TIMEOUT) -n $(NAMESPACE) job/application-gateway-test; \
	echo "Results:"; \
	./scripts/check-pod-logs.sh application-gateway-test; \

gateway-create-resources:
	@echo $(GOPATH)/bin/go-junit-report/v2 --help
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	kubectl label namespace $(NAMESPACE) istio-injection=enabled

	helm template resources/charts/gateway-test \
    --set namespace=$(NAMESPACE) \
    --set mockServiceName=$(MOCK_SERVICE_NAME) \
    | kubectl apply -f -

clean-gateway-test:
	helm template resources/charts/gateway-test  | kubectl delete -f -
	kubectl delete ns $(NAMESPACE) --ignore-not-found

publish-gateway-test: image-gateway-test
	docker push $(GATEWAY_TEST_IMAGE)

image-gateway-test:
	docker build -t $(GATEWAY_TEST_IMAGE) -f Dockerfile.gateway .

publish-mock-app: image-mock-app
	docker push $(MOCK_APP_IMAGE)

image-mock-app:
	docker build -t $(MOCK_APP_IMAGE) -f Dockerfile.mockapp .

disable-sidecar-for-mtls-test:
	kubectl -n kyma-system patch deployment central-application-gateway -p '{"spec":{"template":{"metadata":{"annotations":{"traffic.sidecar.istio.io/excludeOutboundPorts": "8090,8091"}}}}}'
	kubectl rollout status deploy central-application-gateway -n kyma-system --timeout=1m

enable-sidecar-after-mtls-test:
	kubectl -n kyma-system patch deployment central-application-gateway --type=json --patch '[{ "op": "remove", "path": "/spec/template/metadata/annotations/traffic.sidecar.istio.io~1excludeOutboundPorts"}]'

generate-certs:
	./scripts/generate-self-signed-certs.sh $(APP_URL) ./resources/charts/gateway-test/certs/positive
	./scripts/generate-self-signed-certs.sh $(APP_URL) ./resources/charts/gateway-test/certs/negative

patch-for-validator-test:
	kubectl -n kyma-system patch deployment central-application-connectivity-validator --type json --patch-file resources/patches/central-application-connectivity-validator.json
	kubectl rollout status deploy central-application-connectivity-validator -n kyma-system --timeout=1m

unpatch-after-validator-test:
	kubectl rollout undo deployment/central-application-connectivity-validator -n kyma-system

test-validator: patch-for-validator-test validator-create-resources
	if kubectl wait --for=condition=complete --timeout=$(TEST_TIMEOUT) -n $(NAMESPACE) job/application-connectivity-validator-test; then \
	echo "Success! Results:"; \
	./scripts/check-pod-logs.sh application-connectivity-validator-test; \
	$(MAKE) clean-validator-test; \
	else \
	echo "Tests failed! Results:"; \
	./scripts/check-pod-logs.sh application-connectivity-validator-test; \
	$(MAKE) clean-validator-test; \
	exit 1; \
	fi

test-validator-debug: patch-for-validator-test validator-create-resources
	kubectl wait --for=condition=complete --timeout=$(TEST_TIMEOUT) -n $(NAMESPACE) job/application-connectivity-validator-test; \
	echo "Results:"; \
	./scripts/check-pod-logs.sh application-connectivity-validator-test; \

validator-create-resources:
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	kubectl label namespace $(NAMESPACE) istio-injection=enabled --overwrite

	helm template resources/charts/application-connectivity-validator-test/charts/echoserver \
		--set global.namespace=$(NAMESPACE) \
	| kubectl apply -f -
	kubectl rollout status deployment echoserver -n test --timeout=90s

	helm template resources/charts/application-connectivity-validator-test/charts/test \
    		--set namespace=$(NAMESPACE) \
    		--values resources/charts/application-connectivity-validator-test/values.yaml \
	| kubectl apply -f -

clean-validator-test: unpatch-after-validator-test
	helm template resources/charts/application-connectivity-validator-test --set namespace=$(NAMESPACE) | kubectl delete -f -
	kubectl delete ns $(NAMESPACE) --ignore-not-found

publish-validator-test: image-validator-test
	docker push $(VALIDATOR_TEST_IMAGE)

image-validator-test:
	docker build -t $(VALIDATOR_TEST_IMAGE) -f Dockerfile.connectivity-validator .

test-compass-runtime-agent: test-compass-runtime-agent-debug clean-compass-runtime-agent-test

test-compass-runtime-agent-debug:
	@echo $(GOPATH)/bin/go-junit-report --help
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

	helm template resources/charts/compass-runtime-agent-test \
	--set namespace=$(NAMESPACE) \
	--set mockServiceName=$(MOCK_SERVICE_NAME) \
    | kubectl apply -f -

	-kubectl wait --for=condition=complete --timeout=$(TEST_TIMEOUT) -n $(NAMESPACE) job/compass-runtime-agent-test
	@echo ""
	@echo "Compass test results:"

	./scripts/check-pod-logs.sh compass-runtime-agent-test

clean-compass-runtime-agent-test:
	helm template resources/charts/compass-runtime-agent-test  | kubectl delete -f -
	kubectl delete ns $(NAMESPACE) --ignore-not-found

publish-compass-runtime-agent-test: image-compass-runtime-agent-test
	docker push $(COMPASS_TEST_IMAGE)

image-compass-runtime-agent-test:
	docker build -t $(COMPASS_TEST_IMAGE) -f Dockerfile.compass-runtime-agent .

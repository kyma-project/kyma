NAMESPACE ?= test
GOPATH ?= $(shell go env GOPATH)

GATEWAY_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(DOCKER_TAG)"
COMPASS_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/compass-runtime-agent-test:$(DOCKER_TAG)"
MOCK_APP_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/mock-app:$(DOCKER_TAG)"
MOCK_SERVICE_NAME="mock-application"
APP_URL = "$(MOCK_SERVICE_NAME).$(NAMESPACE).svc.cluster.local"
TEST_TIMEOUT = "3m"

OAUTH_COMPASS_CLIENT_ID="aaa"
OAUTH_COMPASS_CLIENT_SECRET="bbb"
OAUTH_COMPASS_TOKENS_ENDPOINT="ccc"


.PHONY: release test image clean
.PHONY: test-gateway test-gateway-debug clean-gateway-test publish-gateway-test image-gateway-test publish-mock-app image-mock-app disable-sidecar-for-mtls-test enable-sidecar-after-mtls-test generate-certs
.PHONY: test-compass-runtime-agent test-compass-runtime-agent-debug clean-compass-runtime-agent-test image-compass-runtime-agent-test publish-compass-runtime-agent-test

release: publish-gateway-test publish-mock-app publish-compass-runtime-agent-test
test: test-gateway test-compass-runtime-agent
image: image-gateway-test image-compass-runtime-agent-test
clean: clean-gateway-test clean-compass-runtime-agent-test

test-gateway: disable-sidecar-for-mtls-test generate-certs create-resources
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


test-gateway-debug: disable-sidecar-for-mtls-test generate-certs create-resources
	kubectl wait --for=condition=complete --timeout=$(TEST_TIMEOUT) -n $(NAMESPACE) job/application-gateway-test; \
	echo "Results:"; \
	./scripts/check-pod-logs.sh application-gateway-test; \

create-resources:
	@echo $(GOPATH)/bin/go-junit-report/v2 --help
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
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

generate-certs:.
	./scripts/generate-self-signed-certs.sh $(APP_URL) ./resources/charts/gateway-test/certs/positive
	./scripts/generate-self-signed-certs.sh $(APP_URL) ./resources/charts/gateway-test/certs/negative

test-compass-runtime-agent: test-compass-runtime-agent-debug clean-compass-runtime-agent-test

test-compass-runtime-agent-debug:
	@echo $(GOPATH)/bin/go-junit-report --help
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

	helm template resources/charts/compass-runtime-agent-test \
    	--set namespace=$(NAMESPACE) \
    	--set mockServiceName=$(MOCK_SERVICE_NAME) \
    	--set compassCredentials.clientID=$(OAUTH_COMPASS_CLIENT_ID) \
    	--set compassCredentials.clientSecret=$(OAUTH_COMPASS_CLIENT_SECRET) \
    	--set compassCredentials.tokensEndpoint=$(OAUTH_COMPASS_TOKENS_ENDPOINT) \
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

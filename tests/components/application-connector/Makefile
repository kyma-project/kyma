NAMESPACE ?= test
GOPATH ?= $(shell go env GOPATH)

GATEWAY_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(DOCKER_TAG)"
COMPASS_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/compass-test:$(DOCKER_TAG)"
MOCK_APP_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/mock-app:$(DOCKER_TAG)"
MOCK_SERVICE_NAME="mock-application"
APP_URL = "$(MOCK_SERVICE_NAME).$(NAMESPACE).svc.cluster.local"
TEST_TIMEOUT = "2m"


.PHONY: release test-os test-skr test image publish clean image-gateway-test publish-gateway-test test-gateway test-gateway-debug test-compass-runtime-agent test-compass-debug image-compass-test clean-compass-test publish-compass-test image-mock-app publish-mock-app disable-sidecar-for-mtls-test enable-sidecar-after-mtls-test generate-certs

test: test-gateway test-compass-runtime-agent
image: image-gateway-test image-compass-test
clean: clean-gateway-test clean-compass-test

release: publish-gateway-test publish-compass-test publish-mock-app

test-gateway: disable-sidecar-for-mtls-test generate-certs test-gateway-debug clean-gateway-test enable-sidecar-after-mtls-test

test-gateway-debug:
	@echo $(GOPATH)/bin/go-junit-report --help
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

	helm template resources/charts/gateway-test \
    	--set namespace=$(NAMESPACE) \
    	--set mockServiceName=$(MOCK_SERVICE_NAME) \
    | kubectl apply -f -

	-kubectl wait --for=condition=complete --timeout=$(TEST_TIMEOUT) -n $(NAMESPACE) job/application-gateway-test
	@echo ""
	@echo "Gateway test results:"

	./scripts/check-pod-logs.sh application-gateway-test

clean-gateway-test:
	helm template resources/charts/gateway-test  | kubectl delete -f -
	kubectl delete ns $(NAMESPACE) --ignore-not-found

publish-gateway-test: image-gateway-test
	docker push $(GATEWAY_TEST_IMAGE)

image-gateway-test:
	docker build -t $(GATEWAY_TEST_IMAGE) -f Dockerfile.gateway .

test-compass-runtime-agent: test-compass-debug clean-compass-test

test-compass-debug:
	@echo $(GOPATH)/bin/go-junit-report --help
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

	helm template resources/charts/compass-test \
    	--set namespace=$(NAMESPACE) \
    	--set mockServiceName=$(MOCK_SERVICE_NAME) \
    | kubectl apply -f -

	-kubectl wait --for=condition=complete --timeout=$(TEST_TIMEOUT) -n $(NAMESPACE) job/compass-runtime-agent-test
	@echo ""
	@echo "Compass test results:"

	./scripts/check-pod-logs.sh compass-runtime-agent-test

clean-compass-test:
	helm template resources/charts/compass-test  | kubectl delete -f -
	kubectl delete ns $(NAMESPACE) --ignore-not-found

publish-compass-test: image-compass-test
	docker push $(COMPASS_TEST_IMAGE)

image-compass-test:
	docker build -t $(COMPASS_TEST_IMAGE) -f Dockerfile.compass .

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



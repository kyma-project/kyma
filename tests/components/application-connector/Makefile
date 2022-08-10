NAMESPACE ?= test
GOPATH ?= $(shell go env GOPATH)

GATEWAY_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(DOCKER_TAG)"
VALIDATOR_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/connectivity-validator-test:$(DOCKER_TAG)"
MOCK_APP_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/mock-app:$(DOCKER_TAG)"
MOCK_SERVICE_NAME="mock-application"
APP_URL = "$(MOCK_SERVICE_NAME).$(NAMESPACE).svc.cluster.local"
TEST_TIMEOUT = "2m"

.PHONY: release test-os test-skr test image publish clean image-gateway-test publish-gateway-test test-gateway test-gateway-debug image-mock-app publish-mock-app disable-sidecar-for-mtls-test enable-sidecar-after-mtls-test generate-certs

test: test-gateway test-validator
image: image-gateway-test image-connectivity-validator-test
clean: clean-gateway-test clean-connectivity-validator-test

release: publish-gateway-test publish-mock-app publish-connectivity-validator-test

test-validator: test-connectivity-validator-debug clean-connectivity-validator-test
test-gateway: disable-sidecar-for-mtls-test generate-certs test-gateway-debug clean-gateway-test enable-sidecar-after-mtls-test


test-connectivity-validator-debug:
	@echo $(GOPATH)/bin/go-junit-report --help
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

	helm template resources/charts/application-connectivity-validator-test | kubectl apply -f -

	kubectl wait --for=condition=complete --timeout=1m -n $(NAMESPACE) job/application-connectivity-validator-test
	@echo ""
	@echo "Application Connectivity Validator test results:"

	./scripts/check-pod-logs.sh application-connectivity-validator-test

clean-connectivity-validator-test:
	helm template resources/charts/application-connectivity-validator-test  | kubectl delete -f -
	kubectl delete ns $(NAMESPACE) --ignore-not-found


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


publish-connectivity-validator-test: image-connectivity-validator-test
	docker push $(VALIDATOR_TEST_IMAGE)

image-connectivity-validator-test:
	docker build -t $(VALIDATOR_TEST_IMAGE) -f Dockerfile.connectivity_validator .


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



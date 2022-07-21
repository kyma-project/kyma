NAMESPACE ?= test
GOPATH ?= $(shell go env GOPATH)

GATEWAY_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(DOCKER_TAG)"
MOCK_APP_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/mock-app:$(DOCKER_TAG)"
GATEWAY_TEST_CERTS_DIR ?= ./resources/charts/gateway-test/certs
APP_URL = "mock-application.test.svc.cluster.local"
SUBJECT =  "/C=PL/ST=A/O=SAP/CN=$(APP_URL)"

.PHONY: release test-os test-skr test image publish clean image-gateway-test publish-gateway-test test-gateway test-gateway-debug image-mock-app publish-mock-app disable-sidecar-for-mtls-test enable-sidecar-after-mtls-test generate-certs

test: disable-sidecar-for-mtls-test generate-certs test-gateway
image: image-gateway-test
clean: clean-gateway-test

release: publish-gateway-test publish-mock-app

test-gateway: test-gateway-debug clean-gateway-test

test-gateway-debug:
	@echo $(GOPATH)/bin/go-junit-report --help
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	helm template resources/charts/gateway-test --set gatewayTest=$(GATEWAY_TEST_IMAGE),mockApplication=$(MOCK_APP_IMAGE),namespace=$(NAMESPACE) | kubectl apply -f -

	-kubectl wait --for=condition=complete --timeout=1m -n $(NAMESPACE) job/application-gateway-test
	@echo ""
	@echo "Gateway test results:"

	./scripts/check-pod-logs.sh

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
	kubectl annotate pod traffic.sidecar.istio.io/excludeOutboundPorts=8090 -l app=central-application-gateway -n kyma-system

enable-sidecar-after-mtls-test:
	kubectl annotate pod traffic.sidecar.istio.io/excludeOutboundPorts- -l app=central-application-gateway -n kyma-system

generate-certs:
	echo "Generating certificate for $(APP_URL)"
	./scripts/generate-self-signed-certs.sh $(APP_URL)

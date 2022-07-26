NAMESPACE ?= test
GOPATH ?= $(shell go env GOPATH)

GATEWAY_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(DOCKER_TAG)"
VALIDATOR_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/app-connectivity-validator-test:$(DOCKER_TAG)"
MOCK_APP_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/mock-app:$(DOCKER_TAG)"

.PHONY: release test image publish clean image-gateway-test publish-gateway-test test-gateway test-gateway-debug image-mock-app publish-mock-app

test: test-gateway test-validator
image: image-gateway-test image-app-connectivity-validator-test
clean: clean-gateway-test clean-app-conn-validator-test

release: publish-gateway-test publish-mock-app publish-app-connectivity-validator-test

test-app-conn-validator: test-app-conn-validator-debug clean-app-conn-validator-test

test-gateway: test-gateway-debug clean-gateway-test


test-app-conn-validator-debug:
	@echo $(GOPATH)/bin/go-junit-report --help
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

clean-app-conn-validator-test:
	helm template resources/charts/application-connectivity-validator-test  | kubectl delete -f -
	kubectl delete ns $(NAMESPACE) --ignore-not-found

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

publish-app-connectivity-validator-test: image-app-connectivity-validator-test
	docker push $(VALIDATOR_TEST_IMAGE)

image-app-connectivity-validator-test:
	docker build -t $(VALIDATOR_TEST_IMAGE) -f Dockerfile.appconnvalidator .

publish-gateway-test: image-gateway-test
	docker push $(GATEWAY_TEST_IMAGE)

image-gateway-test:
	docker build -t $(GATEWAY_TEST_IMAGE) -f Dockerfile.gateway .



publish-mock-app: image-mock-app
	docker push $(MOCK_APP_IMAGE)

image-mock-app:
	docker build -t $(MOCK_APP_IMAGE) -f Dockerfile.mockapp .

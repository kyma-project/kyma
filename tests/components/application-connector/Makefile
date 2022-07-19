NAMESPACE ?= test
GOPATH ?= $(shell go env GOPATH)

GATEWAY_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(DOCKER_TAG)"
MOCK_APP_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/mock-app:$(DOCKER_TAG)"

.PHONY: release test-os test-skr test image publish clean image-gateway-test publish-gateway-test test-gateway test-gateway-debug image-mock-app publish-mock-app

test: test-gateway
image: image-gateway-test
clean: clean-gateway-test

release: publish-gateway-test publish-mock-app

test-gateway: test-gateway-debug clean-gateway-test

test-gateway-debug:
	@echo $(GOPATH)/bin/go-junit-report --help
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	helm template resources/charts/gateway-test --set gatewayTest=$(GATEWAY_TEST_IMAGE),mockApplication=$(MOCK_APP_IMAGE),namespace=$(NAMESPACE) | kubectl apply -f -

	if kubectl wait --for=condition=complete --timeout=1m -n $(NAMESPACE) job/application-gateway-test; then \
	echo "Success! Results:"; \
	./scripts/check-pod-logs.sh; \
	else \
	echo "Tests failed! Results:"; \
	./scripts/check-pod-logs.sh; \
	$(MAKE) clean-gateway-test; \
	exit 1; \
	fi

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

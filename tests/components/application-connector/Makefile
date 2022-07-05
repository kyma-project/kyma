NAMESPACE ?= "test"

DOCKER_TAG ?= "PR-14743"
DOCKER_PUSH_REPOSITORY ?= "eu.gcr.io/kyma-project"

GATEWAY_TEST_IMAGE ?= "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(DOCKER_TAG)"
MOCK_APP_IMAGE ?= "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/mock-app:$(DOCKER_TAG)"

APP_URL = test-app.test.svc.cluster.local

.PHONY: release test-os test image publish clean image-gateway publish-gateway test-gateway certs image-mock-app publish-mock-app

test: test-gateway
image: image-gateway
clean: clean-gateway
publish: publish-gateway publish-mock-app

release: publish
test-os: test-gateway
test-skr: test-gateway

test-gateway:
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	helm template k8s/gateway-test --set gatewayTestImage=$(GATEWAY_TEST_IMAGE),mockImage=$(MOCK_APP_IMAGE),namespace=$(NAMESPACE) | kubectl apply -f -
	sleep 40000
	@echo ""
	@echo ""
	kubectl -n $(NAMESPACE) logs pod/connectivity-test -f | tee /dev/stderr | go-junit-report -set-exit-code > junit-report.xml

clean-gateway:
	helm template k8s/gateway-test --set gatewayTestImage=$(GATEWAY_TEST_IMAGE),mockImage=$(MOCK_APP_IMAGE),namespace=$(NAMESPACE) | kubectl delete -f -

publish-gateway: image-gateway
	docker push $(GATEWAY_TEST_IMAGE)

image-gateway:
	docker build -t $(GATEWAY_TEST_IMAGE) -f Dockerfile.gateway .

publish-mock-app: image-mock-app
	docker push $(MOCK_APP_IMAGE)

image-mock-app:
	docker build -t $(MOCK_APP_IMAGE) -f Dockerfile.testapp .

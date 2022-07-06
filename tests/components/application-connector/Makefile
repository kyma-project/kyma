NAMESPACE ?= test
GOPATH ?= $(shell go env GOPATH)

#for local build & run
DOCKER_TAG ?= "PR-14743"
DOCKER_PUSH_REPOSITORY ?= "eu.gcr.io/kyma-project"
GATEWAY_TEST_IMAGE ?= "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(DOCKER_TAG)"
MOCK_APP_IMAGE ?= "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/mock-app:$(DOCKER_TAG)"

.PHONY: release test-os test image publish clean image-gateway publish-gateway test-gateway certs image-mock-app publish-mock-app

test: test-gateway
image: image-gateway
clean: clean-gateway
publish: publish-gateway publish-mock-app

release: publish
test-os: test-gateway
test-skr: test-gateway

test-gateway:
	@echo $(GOPATH)/bin/go-junit-report --help
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	helm template k8s/gateway-test | kubectl apply -f -
	sleep 180
	@echo ""
	@echo "Gateway test results:"
ifeq ($(EXPORT_RESULT), true)
	kubectl -n $(NAMESPACE) logs pod/application-gateway-test -f | tee /dev/stderr | $(GOPATH)/bin/go-junit-report -set-exit-code > junit-report.xml
else
	kubectl -n $(NAMESPACE) logs pod/application-gateway-test
endif


clean-gateway:
	helm template k8s/gateway-test  | kubectl delete -f -
	kubectl delete ns $(NAMESPACE) --ignore-not-found

publish-gateway: image-gateway
	docker push $(GATEWAY_TEST_IMAGE)

image-gateway:
	docker build -t $(GATEWAY_TEST_IMAGE) -f Dockerfile.gateway .

publish-mock-app: image-mock-app
	docker push $(MOCK_APP_IMAGE)

image-mock-app:
	docker build -t $(MOCK_APP_IMAGE) -f Dockerfile.testapp .

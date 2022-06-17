NAMESPACE ?= "test"
TAG := $(DOCKER_TAG)
TAG ?= "latest"
GATEWAY_IMAGE ?= "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(TAG)"

.PHONY: release test-os test image publish clean image-gateway publish-gateway test-gateway

test: test-gateway
image: image-gateway
clean: clean-gateway
publish: publish-gateway

release: publish
test-os: test-gateway

test-gateway:
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	helm template k8s/gateway-test --set testImage=$(GATEWAY_IMAGE),namespace=$(NAMESPACE) | kubectl apply -f -
	kubectl -n $(NAMESPACE) wait --for=condition=complete job/connectivity-test --timeout 60s

clean-gateway:
	helm template k8s/gateway-test --set testImage=$(GATEWAY_IMAGE) | kubectl delete -f -

publish-gateway: image
	docker push $(GATEWAY_IMAGE)

image-gateway:
	docker build -t $(GATEWAY_IMAGE) -f Dockerfile.gateway .

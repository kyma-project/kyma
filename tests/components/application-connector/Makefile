GATEWAY_IMAGE ?= ""
NAMESPACE ?= "test"

.PHONY: test image publish clean image-gateway publish-gateway test-gateway

test: test-gateway
image: image-gateway
clean: clean-gateway
publish: publish-gateway

test-gateway:
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	helm template k8s/gateway-test --set testImage=$(IMAGE),namespace=$(NAMESPACE) | kubectl apply -f -
	kubectl -n $(NAMESPACE) wait --for=condition=complete job/connectivity-test --timeout 60s

clean-gateway:
	helm template k8s/gateway-test --set testImage=$(IMAGE) | kubectl delete -f -

publish-gateway: image
	docker push $(GATEWAY_IMAGE)

image-gateway:
	docker build -t $(GATEWAY_IMAGE) -f Dockerfile.gateway .

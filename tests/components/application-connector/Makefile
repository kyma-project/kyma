NAMESPACE ?= "test"
TAG := $(DOCKER_TAG)
TAG ?= "latest"
GATEWAY_IMAGE ?= "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(TAG)"
GATEWAY_IMAGE_FOR_TESTS="$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:PR-14504"

APP_URL = test-app.test.svc.cluster.local
OPENSSL = '/usr/local/opt/openssl@3/bin/openssl'

.PHONY: release test-os test image publish clean image-gateway publish-gateway test-gateway certs

test: test-gateway
image: image-gateway
clean: clean-gateway
publish: publish-gateway

release: publish
test-os: test-gateway

test-gateway:
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	helm template k8s/gateway-test --set testImage=$(GATEWAY_IMAGE_FOR_TESTS),namespace=$(NAMESPACE) | kubectl apply -f -
	kubectl -n $(NAMESPACE) wait --for=jsonpath='{.status.phase}'=Running pod/connectivity-test --timeout 60s
	@echo ""
	@echo ""
	kubectl -n $(NAMESPACE) logs pod/connectivity-test -f | tee /dev/stderr | go-junit-report -set-exit-code > junit-report.xml

clean-gateway:
	helm template k8s/gateway-test --set testImage=$(GATEWAY_IMAGE) | kubectl delete -f -

publish-gateway: image
	docker push $(GATEWAY_IMAGE)

image-gateway:
	docker build -t $(GATEWAY_IMAGE) -f Dockerfile.gateway .

certs:
	echo "For app at $(APP_URL)"
	$(OPENSSL) version
	$(OPENSSL) req -new -x509 -nodes -days 365 -subj '/CN=my-ca' -keyout certs/ca.key -out certs/ca.crt

	$(OPENSSL) genrsa -out certs/server.key 2048
	$(OPENSSL) req -new -key certs/server.key -subj '/CN=$(APP_URL)' -out certs/server.csr -addext "subjectAltName = DNS:$(APP_URL)"
	$(OPENSSL) x509 -req -in certs/server.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -days 365 -out certs/server.crt

	$(OPENSSL) genrsa -out certs/client.key 2048
	$(OPENSSL) req -new -key certs/client.key -subj '/CN=my-client' -out certs/client.csr -addext "subjectAltName = DNS:my-client"
	$(OPENSSL) x509 -req -in certs/client.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -days 365 -out certs/client.crt

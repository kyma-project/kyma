NAMESPACE = test
GOPATH ?= $(shell go env GOPATH)

GATEWAY_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(DOCKER_TAG)"
MOCK_APP_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/test-app:$(DOCKER_TAG)"
GATEWAY_TEST_CERTS_DIR ?= ./resources/charts/gateway-test/certs
APP_URL = "mock-app.test.svc.cluster.local"
SUBJECT =  "/C=PL/ST=A/O=SAP/CN=$(APP_URL)"

.PHONY: release test-os test-skr test image publish clean image-gateway-test publish-gateway-test test-gateway image-mock-app publish-mock-app

test: test-gateway
image: image-gateway-test
clean: clean-gateway-test

release: publish-gateway-test publish-mock-app

test-gateway:
	@echo $(GOPATH)/bin/go-junit-report --help
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	helm template resources/charts/gateway-test | kubectl apply -f -

	./scripts/wait-for-gateway-tests-finish.sh

	@echo ""
	@echo "Gateway test results:"
ifeq ($(EXPORT_RESULT), true)
	kubectl -n $(NAMESPACE) logs pod/application-gateway-test -f | tee /dev/stderr | $(GOPATH)/bin/go-junit-report -set-exit-code > junit-report.xml
else
	kubectl -n $(NAMESPACE) logs pod/application-gateway-test
endif


clean-gateway-test:
	helm template k8s/gateway-test  | kubectl delete -f -
	kubectl delete ns $(NAMESPACE) --ignore-not-found

publish-gateway-test: image-gateway-test
	docker push $(GATEWAY_TEST_IMAGE)

image-gateway-test:
	docker build -t $(GATEWAY_TEST_IMAGE) -f Dockerfile.gateway .

publish-mock-app: image-mock-app
	docker push $(MOCK_APP_IMAGE)

image-mock-app:
	docker build -t $(MOCK_APP_IMAGE) -f Dockerfile.mockapp .

certs:
	echo "For app at $(APP_URL)"
	openssl version
	openssl req -newkey rsa:2048 -nodes -x509 -days 365 -out $(GATEWAY_TEST_CERTS_DIR/ca.crt) -keyout $(GATEWAY_TEST_CERTS_DIR/ca.key) -subj $(SUBJECT)
	openssl genrsa -out $(GATEWAY_TEST_CERTS_DIR)/server.key 2048

	openssl req -new \
    -key $(GATEWAY_TEST_CERTS_DIR)/server.key \
    -subj $(SUBJECT) \
    -reqexts SAN \
    -config <(cat /etc/ssl/openssl.cnf \
        <(printf "\n[SAN]\nsubjectAltName=DNS:test-app.test.svc.cluster.local")) \
    -out server.csr
#
#	openssl x509 -req -days 365 -CA ca.crt -CAkey ca.key -CAcreateserial \                                                                                           ✔  11336  12:54:57
#  	-extensions SAN \
#  	-extfile <(cat /etc/ssl/openssl.cnf \
#    <(printf "\n[SAN]\nsubjectAltName=DNS:test-app.test.svc.cluster.local")) \
#  	-in server-ext.csr -out server-clearext.crt

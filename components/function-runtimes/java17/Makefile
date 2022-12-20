#This is address of registry visible from k3s cluster.
#On linux you can get this address by executing `hostname -I`

#TODO: /etc/hosts and coredns patch, fetch it as optional
#REGISTRY_ADR=k3d-kyma-registry.localhost:5000
REGISTRY_ADR=192.168.122.1:5000
JVM_VERSION=17
RUNTIME_IMAGE=java-runtime:${JVM_VERSION}
RUNTIME_IMAGE_REMOTE=${REGISTRY_ADR}/${RUNTIME_IMAGE}

install-sdk:
	(cd serverless-java-sdk && mvn install)

.PHONY: clean
clean:
	mvn clean package

#------------------------------------------------------JVM-------------------------------------------------------------#
.PHONY: build-runtime
build-runtime: install-sdk
	docker build -t ${RUNTIME_IMAGE} -f  Dockerfile-jvm-runtime .

.PHONY: push-runtime-k3d
push-runtime-k3d:
	docker image tag ${RUNTIME_IMAGE}  ${RUNTIME_IMAGE_REMOTE}
	docker push ${RUNTIME_IMAGE_REMOTE}

#------------------------------------------------------K8s-Resources---------------------------------------------------#
#Generate configmaps which can be used to move to serverless resources

IMAGE_HELM_TPL='{{ include "imageurl" (dict "reg" .Values.global.containerRegistry "img" .Values.global.images.function_runtime_java${JVM_VERSION}_jvm_alpha) }}'
.PHONY: generate-configmaps
generate-configmaps:
	DOCKERFILE=Dockerfile-jvm-function.tpl BASE_IMAGE=${IMAGE_HELM_TPL}  python3 ./resources/generate-dockerfile.py  | \
	CONFIGMAP=resources/java${JVM_VERSION}-jvm-alpha.yaml RUNTIME=java${JVM_VERSION}-jvm-alpha python3 resources/generate-cm.py > ./resources/java${JVM_VERSION}-jvm-alpha.yaml

generate-configmaps-local-image:
	DOCKERFILE=Dockerfile-jvm-function.tpl BASE_IMAGE=${RUNTIME_IMAGE_REMOTE}  python3 ./resources/generate-dockerfile.py  | \
	CONFIGMAP=resources/java${JVM_VERSION}-jvm-alpha.yaml RUNTIME=java${JVM_VERSION}-jvm-alpha python3 resources/generate-cm.py > ./resources/java${JVM_VERSION}-jvm-alpha-local.yaml


apply-java-runtime:
	kubectl replace -f ./resources/java17-jvm-alpha.yaml

apply-java-runtime-local:
	kubectl replace -f ./resources/java17-jvm-alpha-local.yaml

#-------------------------------------------------Run Example function locally-----------------------------------------#
#Example function
#Create Dockerfile to use with example hello-world
export BASE_IMAGE = ${RUNTIME_IMAGE}

FUNCTION_IMAGE=java-jvm${JVM_VERSION}-function

.PHONY: generate-local-fn-dockerfile
generate-local-fn-dockerfile:
	DOCKERFILE=Dockerfile-jvm-function.tpl BASE_IMAGE=${RUNTIME_IMAGE}  python3 ./resources/generate-dockerfile.py > Dockerfile-jvm-function-local

run-jvm-hello-world: build-runtime generate-local-fn-dockerfile
	docker build --tag ${FUNCTION_IMAGE} \
		  --build-arg DEPS_DIR=./examples/hello-world/ \
		 --build-arg SOURCE_DIR=./examples/hello-world/src/main/java/io/project/kyma/serverless/handler \
		 -f Dockerfile-jvm-function-local .

	docker run -ti -p 8080:8080 --rm --env PUBLISHER_PROXY_ADDRESS=http://localhost:10000/publish \
		--env TRACE_COLLECTOR_ENDPOINT=http://localhost:4318/v1/traces \
		--env SERVICE_NAMESPACE=default \
		--env HOSTNAME=emitter-local-84dd76fc94-2pnpd \
		--name java-jvm-function ${FUNCTION_IMAGE}

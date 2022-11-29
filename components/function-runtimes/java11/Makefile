REGISTRY_ADR=192.168.122.1:5000
RUNTIME_IMAGE=java-runtime:11
RUNTIME_IMAGE_REMOTE=${REGISTRY_ADR}/${RUNTIME_IMAGE}

install-sdk:
	(cd serverless-java-sdk && mvn install)

.PHONY: clean
clean:
	mvn clean package

#------------------------------------------------------JVM---------------------------------------------------------#
.PHONY: build-runtime
build-runtime: install-sdk
	docker build -t ${RUNTIME_IMAGE} -f  Dockerfile-11jvm-runtime .

.PHONY: build-runtime
push-runtime-k3d:
	docker image tag ${RUNTIME_IMAGE}  ${RUNTIME_IMAGE_REMOTE}
	docker push ${RUNTIME_IMAGE_REMOTE}


#------------------------------------------------------GraalVM---------------------------------------------------------#
#TODO: No need to build graalvm
.PHONY: build-graalvm-base-image
build-graalvm:
	docker build -t graal-vm:11 -f Dockerfile-graalvm-11 .

.PHONY: build-graalvm-runtime
build-runtime-graalvm:
	docker build -t graal-vm-runtime:11 -f Dockerfile-graalvm-runtime .


push-runtime-graalvm:
	docker tag graal-vm-runtime:11 "${REGISTRY_ADR}"/graal-vm-runtime:11
	docker push "${REGISTRY_ADR}"/graal-vm-runtime:11

#This step generate files which are used by native-image to include all needed classes which are dynamically loaded at runtime
#Run this and make some calls to function
#https://www.graalvm.org/22.3/reference-manual/native-image/guides/build-with-reflection/
.PHONY: generate-reflection-native
generate-reflection-native: clean
	OUTPUT_PATH="./src/main/resouces/META-INF/native-image/io.project.kyma.serverless"
	mkdir -p "${OUTPUT_PATH}"
	java -agentlib:native-image-agent=config-output-dir="${OUTPUT_PATH}"  -jar target/kyma-java-runtime-0.0.1.jar --native


#------------------------------------------------------K8s-Resources---------------------------------------------------#
#Generate configmaps which can be used to move to serverless resources

IMAGE_HELM_TPL='{{ include "imageurl" (dict "reg" .Values.global.containerRegistry "img" .Values.global.images.function_runtime_java11-jvm-alpha) }}'
.PHONY: generate-configmaps
generate-configmaps:
	DOCKERFILE=Dockerfile-jvm-function.tpl BASE_IMAGE=${IMAGE_HELM_TPL} CONFIGMAP=resources/java11-jvm-alpha.yaml  python3 resources/generate-cm.py


apply-java-runtime:
	kubectl replace -f ./resources/java11-jvm-alpha.yaml

#-------------------------------------------------Run Example function locally-----------------------------------------#
#Example function
#Create Dockerfile to use with example hello-world
export BASE_IMAGE = ${RUNTIME_IMAGE}

FUNCTION_IMAGE=java-jvm11-function

.PHONY: Dockerfile-jvm11-function-local
Dockerfile-jvm11-function-local:
	DOCKERFILE=Dockerfile-jvm-function.tpl BASE_IMAGE=${RUNTIME_IMAGE_REMOTE}  python3 ./resources/generate-dockerfile.py

run-jvm-hello-world: build-runtime Dockerfile-jvm11-function-local
	docker build --tag ${FUNCTION_IMAGE} \
		  --build-arg DEPS_DIR=./examples/hello-world/ \
		 --build-arg SOURCE_DIR=./examples/hello-world/src/main/java/io/project/kyma/serverless/handler \
		 -f Dockerfile-jvm11-function-local .

	docker run -ti -p 8080:8080 --rm --env PUBLISHER_PROXY_ADDRESS=http://localhost:10000/publish \
		--env TRACE_COLLECTOR_ENDPOINT=http://localhost:4318/v1/traces\
		--env SERVICE_NAMESPACE=default \
		--env HOSTNAME=emitter-local-84dd76fc94-2pnpd \
		--name java-jvm11-function ${FUNCTION_IMAGE}

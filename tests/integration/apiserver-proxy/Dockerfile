FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1 as builder

ENV BASE_DIR /workspace/go/src/github.com/kyma-project/kyma/tests/integration/apiserver-proxy
WORKDIR ${BASE_DIR}

COPY ./fetch-token/ ${BASE_DIR}/fetch-token/
COPY ./licenses/ ${BASE_DIR}/licenses/
COPY ./vendor/ ${BASE_DIR}/vendor/
COPY ./test.sh ${BASE_DIR}/test.sh

##
# BUILD FETCH-TOKEN APP
##
RUN CGO_ENABLED=0 GOOS=linux go build -v -o ${BASE_DIR}/app ${BASE_DIR}/fetch-token/main.go

##
# EXTRACT SCRIPT AND BINARY
##
RUN mkdir /test && mv ${BASE_DIR}/app /test/app && mv ${BASE_DIR}/test.sh /test/test.sh && mv ${BASE_DIR}/licenses /test/licenses

FROM eu.gcr.io/kyma-project/tpi/k8s-tools:20210504-12243229
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder --chown=nobody:nogroup /test /test
WORKDIR /test

RUN apk update &&\
	apk add curl bash grep &&\
	curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl &&\
	chmod +x kubectl &&\
	mv kubectl /usr/local/bin/kubectl &&\
	chown -R nobody:nogroup /test

USER nobody

ENTRYPOINT ["/bin/bash", "-c", "/test/test.sh"]

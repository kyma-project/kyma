FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1 as builder

ENV BASE_DIR /workspace/go/src/github.com/kyma-project/kyma/tests/integration/api-gateway
WORKDIR ${BASE_DIR}

COPY . ${BASE_DIR}/
COPY ./licenses/ ${BASE_DIR}/licenses/

##
# BUILD APP
##
RUN cd ${BASE_DIR}/gateway-tests &&\
	GO111MODULE=on go mod vendor &&\
	CGO_ENABLED=0 GOOS=linux go test -c -v -o ${BASE_DIR}/app.test main_test.go

##
# EXTRACT SCRIPT AND BINARY
##
RUN mkdir /test &&\
	mv ${BASE_DIR}/app.test /test/app.test  &&\
	mv ${BASE_DIR}/licenses /test/licenses

FROM eu.gcr.io/kyma-project/tpi/k8s-tools:20210504-12243229
LABEL source=git@github.com:kyma-project/kyma.git
WORKDIR /test

RUN apk update &&\
	apk add curl bash grep ca-certificates

COPY --from=builder --chown=nobody:nogroup /test /test
COPY gateway-tests/manifests /test/manifests

USER nobody
ENTRYPOINT ["/test/app.test"]

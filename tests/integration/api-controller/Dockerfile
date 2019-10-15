FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1 as builder

ENV SRC_DIR /workspace/go/src/github.com/kyma-project/kyma/tests/integration/api-controller/

WORKDIR ${SRC_DIR}

#
# Copy files
#

COPY ./apicontroller ${SRC_DIR}/apicontroller/
COPY ./vendor ${SRC_DIR}/vendor/
COPY ./licenses/ ${BASE_APP_DIR}/licenses/

#
# Build app
#

RUN CGO_ENABLED=0 go test -c ./apicontroller/ -o /apicontroller.test
RUN mkdir /test && mv /apicontroller.test /test/apicontroller.test && mv ${BASE_APP_DIR}/licenses /test/licenses

FROM alpine:3.10

LABEL source = git@github.com:kyma-project/kyma.git

#
# Install certificates
#

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

#
# Copy binary + licences
#

COPY --from=builder /test /test

#
# Run tests
#

ENTRYPOINT [ "/test/apicontroller.test", "-test.v" ]

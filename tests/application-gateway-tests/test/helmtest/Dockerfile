FROM golang:1.10-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-gateway-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN go test -o proxyhelmtests.test -c ./test/helmtest/proxy/tests

FROM alpine:3.8

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/scripts/helm-test-entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/proxyhelmtests.test .

ARG TEST_EXECUTOR_IMAGE
ENV TEST_EXECUTOR_IMAGE=$TEST_EXECUTOR_IMAGE

ENTRYPOINT ./helm-test-entrypoint.sh

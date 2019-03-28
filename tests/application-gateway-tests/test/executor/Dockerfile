FROM golang:1.10-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-gateway-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN go test -o proxytestsexecutor.test -c ./test/executor/proxy/tests

FROM alpine:3.8

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/scripts/executor-entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/proxytestsexecutor.test .

ENTRYPOINT ./executor-entrypoint.sh

FROM golang:1.13-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-gateway-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 go test -o proxytestsexecutor.test -c ./test/executor/proxy/tests

FROM alpine:3.10

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/spec.json .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/scripts/executor-entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/proxytestsexecutor.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/licenses ./licenses

ENTRYPOINT ./executor-entrypoint.sh

FROM golang:1.13-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/connection-token-handler-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 go test -c ./test/k8stests

FROM alpine:3.10

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/connection-token-handler-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/connection-token-handler-tests/k8stests.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/connection-token-handler-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh

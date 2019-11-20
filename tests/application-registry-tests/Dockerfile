FROM golang:1.13-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-registry-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 go test -c ./test/apitests
RUN CGO_ENABLED=0 go test -c ./test/k8stests

FROM alpine:3.10

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-registry-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-registry-tests/apitests.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-registry-tests/k8stests.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-registry-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh

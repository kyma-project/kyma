FROM golang:1.9-alpine3.7 as builder
LABEL source=git@github.com:kyma-project/kyma.git

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/metadata-service-tests

ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN go test -c ./test/apitests
RUN go test -c ./test/k8stests

FROM alpine

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/metadata-service-tests .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/metadata-service-tests/scripts/entrypoint.sh .

ENTRYPOINT ./entrypoint.sh

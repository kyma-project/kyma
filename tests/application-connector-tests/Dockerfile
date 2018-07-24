FROM golang:1.9-alpine3.7
LABEL source=git@github.com:kyma-project/kyma.git

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-connector-tests

ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN go test -c ./test/metadata/apitests
RUN go test -c ./test/metadata/k8stests

ENTRYPOINT ./scripts/entrypoint.sh

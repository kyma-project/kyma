FROM golang:1.9-alpine3.7

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/gateway-tests

ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN go test -c ./test/apitests

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ./entrypoint.sh

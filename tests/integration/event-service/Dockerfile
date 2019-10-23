FROM golang:1.12.10-alpine3.10

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/integration/event-service

ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN go test -c ./test/apitests

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ./entrypoint.sh

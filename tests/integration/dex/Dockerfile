FROM golang:1.10-alpine3.7

RUN apk update && apk add curl

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/integration/dex
ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN go test -c ./dex

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ./entrypoint.sh

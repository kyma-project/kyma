FROM golang:1.10-alpine3.7

RUN apk update && apk add curl

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/acceptance
ADD . $SRC_DIR

ADD client.bin /go/bin/client.bin
ADD gateway.bin /go/bin/gateway.bin
ADD env-tester.bin /go/bin/env-tester.bin

WORKDIR $SRC_DIR

RUN go test -c ./application
RUN go test -c ./servicecatalog
RUN go test -c ./dex

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ./entrypoint.sh

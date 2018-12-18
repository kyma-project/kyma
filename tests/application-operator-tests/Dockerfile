FROM golang:1.9-alpine3.7 as builder
LABEL source=git@github.com:kyma-project/kyma.git

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-operator-tests

ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN go test -o controllertests.test -c ./test/controllertests

FROM alpine

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/scripts/entrypoint.sh .

ENTRYPOINT ./entrypoint.sh

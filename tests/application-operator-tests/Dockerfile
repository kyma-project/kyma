FROM golang:1.10-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-operator-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN go test -o controllertests.test -c ./test/controllertests

FROM alpine:3.8

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/controllertests.test .

ENTRYPOINT ./entrypoint.sh

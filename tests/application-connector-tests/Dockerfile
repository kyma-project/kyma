FROM golang:1.10-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-connector-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN go test -o applicationaccess.test -c ./test/applicationaccess/tests

FROM alpine:3.8

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-connector-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-connector-tests/applicationaccess.test .

ENTRYPOINT ./entrypoint.sh

FROM golang:1.10-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/connector-service-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN go test -c ./test/apitests

FROM alpine:3.8

LABEL source=git@github.com:kyma-project/kyma.git

RUN apk --no-cache add ca-certificates

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/connector-service-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/connector-service-tests/apitests.test .

ENTRYPOINT ./entrypoint.sh

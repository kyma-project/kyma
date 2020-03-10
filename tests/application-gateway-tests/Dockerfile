FROM golang:1.13-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-gateway-tests

RUN apk add -U --no-cache ca-certificates

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 GOOS=linux go test -c ./test/gateway/tests

FROM alpine:3.10

LABEL source=git@github.com:kyma-project/kyma.git
RUN apk add --no-cache curl

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/tests.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh

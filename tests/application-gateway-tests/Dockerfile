FROM eu.gcr.io/kyma-project/external/golang:1.14.8-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-gateway-tests

RUN apk add -U --no-cache ca-certificates
WORKDIR $SRC_DIR

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . $SRC_DIR
COPY ./licenses/ /app/licenses

RUN CGO_ENABLED=0 GOOS=linux go test -c ./test/gateway/tests

FROM eu.gcr.io/kyma-project/external/alpine:3.12.0

LABEL source=git@github.com:kyma-project/kyma.git
RUN apk add --no-cache curl

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/tests.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-gateway-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh

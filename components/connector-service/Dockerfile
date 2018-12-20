FROM golang:1.10-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/connector-service

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o connectorservice ./cmd/connectorservice

FROM alpine:3.8
LABEL source = git@github.com:kyma-project/kyma.git

RUN apk --no-cache add curl ngrep openssl

COPY ./docs/api/connectorapi.yaml .
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/connector-service/connectorservice .

CMD ["/connectorservice"]

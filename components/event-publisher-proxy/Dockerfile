FROM golang:1.15-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/event-publisher-proxy

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN GOOS=linux GO111MODULE=on go mod vendor && \
    CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -o event-publisher-proxy ./cmd/event-publisher-proxy

FROM scratch
LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/event-publisher-proxy/event-publisher-proxy .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY licenses/ /licenses/

ENTRYPOINT ["/event-publisher-proxy"]

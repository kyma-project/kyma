FROM golang:1.15-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/cloud-event-gateway-proxy

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cloud-event-gateway-proxy ./cmd/cloud-event-gateway-proxy

FROM scratch
LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/cloud-event-gateway-proxy .
COPY licenses/ /licenses/

ENTRYPOINT ["/cloud-event-gateway-proxy"]

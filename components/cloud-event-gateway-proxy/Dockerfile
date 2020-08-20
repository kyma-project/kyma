FROM golang:1.14-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/cloud-event-gateway-proxy

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR
RUN ls -lrt ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cloud-event-gateway-proxy ./cmd

FROM shazra/nginx
LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/cloud-event-gateway-proxy .

CMD ["/cloud-event-gateway-proxy"]
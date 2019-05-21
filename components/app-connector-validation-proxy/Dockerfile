FROM golang:1.12-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/app-connector-validation-proxy

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o appconnectorvalidationproxy ./cmd/appconnectorvalidationproxy

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/app-connector-validation-proxy .

CMD ["/appconnectorvalidationproxy"]

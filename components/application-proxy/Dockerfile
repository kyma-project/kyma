FROM golang:1.10-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/application-proxy

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o applicationproxy ./cmd/applicationproxy

FROM alpine:3.8 as certs

RUN apk add -U --no-cache ca-certificates

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-proxy .

CMD ["/applicationproxy"]

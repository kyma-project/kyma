FROM golang:1.10-alpine as builder

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/application-gateway

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR
COPY ./licenses/ /app/licenses

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o applicationgateway ./cmd/applicationgateway

FROM alpine:3.10 as certs

RUN apk add -U --no-cache ca-certificates

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/application-gateway .
COPY --from=builder /app/licenses /app/licenses 

CMD ["/applicationgateway"]

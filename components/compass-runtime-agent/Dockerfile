FROM golang:1.12.5-alpine3.9 as builder

RUN apk add -U --no-cache ca-certificates

ARG DOCK_PKG_DIR=/go/src/github.com/kyma-project/kyma/components/compass-runtime-agent

WORKDIR $DOCK_PKG_DIR
COPY . $DOCK_PKG_DIR
COPY ./licenses/ /app/licenses

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o compass-runtime-agent ./cmd

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /go/src/github.com/kyma-project/kyma/components/compass-runtime-agent/compass-runtime-agent .
COPY --from=builder /app/licenses /app/licenses

CMD ["/compass-runtime-agent"]
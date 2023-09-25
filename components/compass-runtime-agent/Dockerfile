# image builder base on golang:1.21.1-alpine3.18
FROM golang@sha256:0c860c7ceba62231d0f99fb92e9d7c1577f26fea794a12c75756a8f64b146e45 as builder

ARG DOCK_PKG_DIR=/compass-runtime-agent
WORKDIR $DOCK_PKG_DIR

COPY . $DOCK_PKG_DIR

RUN echo "nobody:x:65534:5534:nobody:/:" > /etc_passwd

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o compass-runtime-agent ./cmd


FROM eu.gcr.io/kyma-project/external/alpine:3.17.3 as certs
RUN apk add -U --no-cache ca-certificates

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

WORKDIR /app

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder $DOCK_PKG_DIR/compass-runtime-agent .

COPY --from=builder /etc_passwd /etc/passwd
USER nobody

CMD ["/app/compass-runtime-agent"]

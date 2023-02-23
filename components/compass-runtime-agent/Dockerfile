FROM eu.gcr.io/kyma-project/external/golang:1.20.1-alpine3.17 as builder

ARG DOCK_PKG_DIR=/compass-runtime-agent
WORKDIR $DOCK_PKG_DIR

COPY . $DOCK_PKG_DIR

RUN echo "nobody:x:65534:5534:nobody:/:" > /etc_passwd

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o compass-runtime-agent ./cmd


FROM eu.gcr.io/kyma-project/external/alpine:3.17.2 as certs
RUN apk add -U --no-cache ca-certificates

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

WORKDIR /app

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder $DOCK_PKG_DIR/compass-runtime-agent .

COPY --from=builder /etc_passwd /etc/passwd
USER nobody

CMD ["/app/compass-runtime-agent"]

FROM eu.gcr.io/kyma-project/external/golang:1.16.0-alpine as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/kyma/components/function-controller \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR ${BASE_APP_DIR}

# Copy the go source
COPY . ${BASE_APP_DIR}/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o webhook-server cmd/webhook/main.go \
&& mkdir /app \
&& mv ./webhook-server /app/webhook-server

FROM alpine:3.13.5 as certs
RUN apk --update add ca-certificates

FROM scratch

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /app /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 1000

ENTRYPOINT ["/app/webhook-server"]

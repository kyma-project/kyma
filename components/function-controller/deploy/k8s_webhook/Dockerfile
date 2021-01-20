FROM golang:1.15-alpine as builder

ENV BASE_APP_DIR=/go/src/github.com/kyma-project/kyma/components/function-controller \
    GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR ${BASE_APP_DIR}

COPY ./go.mod .
COPY ./go.sum .

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . ${BASE_APP_DIR}/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o k8s-webhook-server cmd/k8s_webhook/main.go

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch

LABEL source = git@github.com:kyma-project/kyma.git

WORKDIR /
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/function-controller/k8s-webhook-server .
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 1000

ENTRYPOINT ["/k8s-webhook-server"]

FROM golang:1.13 as builder

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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o webhook-server cmd/webhook/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:3.11
WORKDIR /
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/function-controller/webhook-server .
USER 1000

ENTRYPOINT ["/webhook-server"]

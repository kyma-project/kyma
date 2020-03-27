# Build the manager binary
FROM golang:1.13 as builder
#
# Copy in the go src
#
WORKDIR /go/src/github.com/kyma-project/kyma/components/function-controller

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY migrator/pkg/    migrator/pkg/
COPY migrator/cmd/    migrator/cmd/
COPY pkg/apis         pkg/apis
#
# Build
#
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o main migrator/cmd/main.go

FROM alpine:3.11
LABEL source=git@github.com:kyma-project/kyma.git
#
# Copy binary
#

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*


COPY --from=builder /go/src/github.com/kyma-project/kyma/components/function-controller/main .
#
# Run app
#
ENTRYPOINT ["/main"]
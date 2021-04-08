# Build the manager binary
FROM eu.gcr.io/kyma-project/external/golang:1.16.0-alpine as builder

WORKDIR /workspace
# Copy the go source
COPY go.mod .
COPY go.sum .
COPY main.go main.go
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]

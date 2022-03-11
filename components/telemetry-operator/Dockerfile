# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

FROM eu.gcr.io/kyma-project/tpi/fluent-bit:1.8.13-bdc43fc8 as fluentbit


# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM debian:testing-20220228-slim
RUN apt-get update && \
    apt-get -y dist-upgrade && \
    apt-get install -y --no-install-recommends \
    libsasl2-2 \
    libssl1.1 \
    ca-certificates && \
    rm -rf /var/lib/apt/lists
WORKDIR /

COPY --from=fluentbit /usr/lib/x86_64-linux-gnu/libpq.so* /usr/lib/x86_64-linux-gnu/
COPY --from=fluentbit /usr/lib/x86_64-linux-gnu/libldap* /usr/lib/x86_64-linux-gnu/
COPY --from=fluentbit /usr/lib/x86_64-linux-gnu/liblber* /usr/lib/x86_64-linux-gnu/


COPY --from=builder /workspace/manager .
COPY --from=fluentbit /fluent-bit .
USER 65532:65532

ENTRYPOINT ["/manager"]

FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20210402-70b4b74f as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/components/console-backend-service
ENV GO111MODULE on

WORKDIR ${BASE_APP_DIR}

COPY ./go.mod ./go.sum ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY . .

# Run tests
RUN make verify

# Build app
RUN make build

# Copy built app
RUN mkdir /app
RUN mv ./console-backend-service /app/main 

COPY ./licenses/ /app/licenses

# Install certificates
FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch

LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /app

# Copy binary
COPY --from=builder /app /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# Run app

# -logtostderr=true is required, as glog is stupid and tries to log to /tmp, which crashes the pod
CMD ["/app/main", "-logtostderr=true"]

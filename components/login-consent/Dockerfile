FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20200117-d3885041 as builder

ENV BASE_APP_DIR /workspace/go/src/github.com/kyma-project/kyma/components/login-consent
WORKDIR ${BASE_APP_DIR}

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY main.go main.go
COPY internal/ internal/

#
# Build app
#

RUN CGO_ENABLED=0 GOOS=linux go build -v -o app ${BASE_APP_DIR}/main.go
RUN mkdir /app && mv ./app /app/app
#&& mv ${BASE_APP_DIR}/licenses /app/licenses

FROM alpine:3.12
LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /app

#
# Install certificates
#

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

#
# Copy binary
#

COPY --from=builder /app /app

#
# Run app
#

ENTRYPOINT ["/app/app"]

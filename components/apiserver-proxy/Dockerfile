FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20200117-d3885041 as builder

ENV BASE_APP_DIR /workspace/go/src/github.com/kyma-project/kyma/components/apiserver-proxy
WORKDIR ${BASE_APP_DIR}


#
# Copy files
#

COPY ./cmd/ ${BASE_APP_DIR}/cmd/
COPY ./internal/ ${BASE_APP_DIR}/internal/
COPY ./licenses/ ${BASE_APP_DIR}/licenses/
COPY ./vendor/ ${BASE_APP_DIR}/vendor/

#
# Build app
#

RUN CGO_ENABLED=0 GOOS=linux go build -v -o app ${BASE_APP_DIR}/cmd/proxy/main.go
RUN mkdir /app && mv ./app /app/app && mv ${BASE_APP_DIR}/licenses /app/licenses

FROM alpine:3.11
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

FROM eu.gcr.io/kyma-project/external/golang:1.16.0-alpine as builder

ENV BASE_APP_DIR /workspace/go/src/github.com/kyma-project/kyma/components/apiserver-proxy
WORKDIR ${BASE_APP_DIR}


#
# Copy files
#

COPY . ${BASE_APP_DIR}

#
# Build app
#

RUN CGO_ENABLED=0 GOOS=linux go build -v -o app ${BASE_APP_DIR}/cmd/proxy/main.go
RUN mkdir /app && mv ./app /app/app && mv ${BASE_APP_DIR}/licenses /app/licenses

FROM alpine:3.13.5
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

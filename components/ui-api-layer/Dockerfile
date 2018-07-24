FROM golang:1.10.2-alpine3.7 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/components/ui-api-layer
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./internal/ ${BASE_APP_DIR}/internal/
COPY ./pkg/ ${BASE_APP_DIR}/pkg/
COPY ./main.go ${BASE_APP_DIR}/

#
# Build app
#

RUN go build -v -o main .
RUN mkdir /app && mv ./main /app/main

FROM alpine:3.7
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

CMD ["/app/main"]

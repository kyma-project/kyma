FROM golang:1.11.4-alpine3.8 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/components/console-backend-service
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./internal/ ${BASE_APP_DIR}/internal/
COPY ./pkg/ ${BASE_APP_DIR}/pkg/
COPY ./main.go ${BASE_APP_DIR}/
COPY ./licenses/ ${BASE_APP_DIR}/licenses/

#
# Build app
#

RUN CGO_ENABLED=0 GOOS=linux go build -v -o main .
RUN mkdir /app && mv ./main /app/main && mv ${BASE_APP_DIR}/licenses /app/licenses

FROM alpine:3.10
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

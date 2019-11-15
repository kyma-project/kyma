FROM golang:1.11.4-alpine3.8 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/components/asset-upload-service
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

RUN CGO_ENABLED=0 go build -v -o main .
RUN mkdir /app && mv ./main /app/main && mv ${BASE_APP_DIR}/licenses /app/licenses

# Get latest CA certs
FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch

LABEL source = git@github.com:kyma-project/kyma.git

WORKDIR /app

COPY --from=builder /app /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/app/main","-logtostderr","=false"]

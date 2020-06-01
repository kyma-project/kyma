FROM golang:1.13.11-alpine3.11 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/components/console-backend-service

ENV GO111MODULE on

WORKDIR ${BASE_APP_DIR}


COPY ./go.mod .
COPY ./go.sum .

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

#

#
# Copy files
#

COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./internal/ ${BASE_APP_DIR}/internal/
COPY ./pkg/ ${BASE_APP_DIR}/pkg/
COPY ./main.go ${BASE_APP_DIR}/
COPY ./licenses/ ${BASE_APP_DIR}/licenses/
# COPY . ${BASE_APP_DIR}/

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

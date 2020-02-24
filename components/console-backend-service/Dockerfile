FROM golang:1.13 as builder

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

FROM alpine:3.10 as certs 
RUN apk add -U --no-cache ca-certificates

FROM scratch
LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kyma-project/kyma/components/console-backend-service .
COPY --from=builder /app/licenses /app/licenses 

CMD ["/app/main"]

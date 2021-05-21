FROM golang:1.15-alpine3.12 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/tests/rafter
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./testsuite/ ${BASE_APP_DIR}/testsuite/
COPY ./pkg/ ${BASE_APP_DIR}/pkg/
COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./main_test.go ${BASE_APP_DIR}
COPY ./licenses/ ${BASE_APP_DIR}/licenses/

#
# Build tests
#

RUN CGO_ENABLED=0 go test main_test.go -c

FROM alpine:3.13.5
LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /app

ENV TEST_DIR /go/src/github.com/kyma-project/kyma/tests/rafter

#
# Install certificates and tools
#

RUN apk update --no-cache && \
apk upgrade --no-cache && \
apk add ca-certificates && \
apk add --no-cache curl>=7.76.0-r0 --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main && \
rm -rf /var/cache/apk/*

#
# Copy binaries
#

RUN mkdir -p /app/tests
COPY --from=builder ${TEST_DIR}/main.test /app/
COPY --from=builder ${TEST_DIR}/licenses/ /app/licenses/

#
# Run tests
#
USER 1000

CMD ["/app/main.test"]

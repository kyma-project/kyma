FROM golang:1.17.6-alpine3.15 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/tests/rafter
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./testsuite/ ${BASE_APP_DIR}/testsuite/
COPY ./pkg/ ${BASE_APP_DIR}/pkg/
COPY ./main_test.go ${BASE_APP_DIR}
COPY ./go.mod ${BASE_APP_DIR}
COPY ./go.sum ${BASE_APP_DIR}
COPY ./licenses/ ${BASE_APP_DIR}/licenses/

#
# Build tests
#

RUN CGO_ENABLED=0 go test main_test.go -c

FROM alpine:3.14.2
LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /app

ENV TEST_DIR /go/src/github.com/kyma-project/kyma/tests/rafter

#
# Install certificates and tools
#

RUN apk upgrade --no-cache && \
    apk add --no-cache ca-certificates && \
    apk add --no-cache curl>=7.77.0-r1 --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main

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

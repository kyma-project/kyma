FROM golang:1.15-alpine3.12 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/tests/function-controller
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./go.mod ${BASE_APP_DIR}
COPY ./go.sum ${BASE_APP_DIR}

RUN GO111MODULE=on go mod download


COPY ./testsuite/ ${BASE_APP_DIR}/testsuite/
COPY ./pkg/ ${BASE_APP_DIR}/pkg/
COPY ./main_test.go ${BASE_APP_DIR}
COPY ./licenses/ ${BASE_APP_DIR}/licenses/

#
# Build tests
#

RUN CGO_ENABLED=0 go test main_test.go -c

FROM alpine:3.12
LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /app

ENV TEST_DIR /go/src/github.com/kyma-project/kyma/tests/function-controller

#
# Install certificates and tools
#

RUN apk update && apk add ca-certificates && apk add curl && rm -rf /var/cache/apk/*

#
# Copy binaries
#

COPY --from=builder ${TEST_DIR}/main.test /app/
COPY --from=builder ${TEST_DIR}/licenses/ /app/licenses/

#
# Run tests
#

CMD ["/app/main.test"]

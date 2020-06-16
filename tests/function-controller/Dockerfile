FROM golang:1.14-alpine3.11 as builder

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

FROM alpine:latest as certs
RUN apk --update add ca-certificates

LABEL source = git@github.com:kyma-project/kyma.git

ENV TEST_DIR /go/src/github.com/kyma-project/kyma/tests/function-controller

FROM scratch

WORKDIR /app

ENV TEST_DIR /go/src/github.com/kyma-project/kyma/tests/function-controller

COPY --from=builder ${TEST_DIR}/main.test /app/
COPY --from=builder ${TEST_DIR}/licenses/ /app/licenses/
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

CMD ["/app/main.test"]

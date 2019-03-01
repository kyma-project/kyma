FROM golang:1.11.4-alpine3.8 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/tests/asset-store
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./internal/ ${BASE_APP_DIR}/internal/
COPY ./pkg/ ${BASE_APP_DIR}/pkg/
COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./tester_test.go ${BASE_APP_DIR}

#
# Build tests
#

RUN CGO_ENABLED=0 go test tester_test.go -c

FROM alpine:3.8
LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /app

ENV TEST_DIR /go/src/github.com/kyma-project/kyma/tests/asset-store/main.test

#
# Install certificates
#

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

#
# Copy binaries
#

RUN mkdir -p /app/tests
COPY --from=builder ${TEST_DIR} /app/tests/

#
# Run tests
#

CMD /app/tests/main.test

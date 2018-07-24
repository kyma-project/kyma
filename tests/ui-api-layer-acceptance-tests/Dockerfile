FROM golang:1.10.2-alpine3.7 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./brokerinstaller/ ${BASE_APP_DIR}/brokerinstaller/
COPY ./domain/ ${BASE_APP_DIR}/domain/
COPY ./graphql/ ${BASE_APP_DIR}/graphql/
COPY ./k8s/ ${BASE_APP_DIR}/k8s/
COPY ./waiter/ ${BASE_APP_DIR}/waiter/
COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./tester.go ${BASE_APP_DIR}/tester.go

#
# Build tests
#

RUN mkdir -p /app
RUN cd ${BASE_APP_DIR}/domain && \
    for dir in $(ls -1FA | grep / | tr -d /); do \
    echo "Building test package ${dir}..."; \
    go test -tags=acceptance ./${dir} -c \
    ; done

FROM alpine:3.7
LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /app

ENV TESTS_DIR /go/src/github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/domain
ENV TEST_SERVICE_CATALOG_DIR /app/tests/servicecatalog

#
# Install certificates
#

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

#
# Copy binaries
#

RUN mkdir -p /app/tests
COPY --from=builder ${TESTS_DIR} /app/tests/

#
# Run tests
#

CMD cd /app/tests/ && \
        for f in *.test; do \
        chmod +x ./${f} && ./${f} \
        ; done


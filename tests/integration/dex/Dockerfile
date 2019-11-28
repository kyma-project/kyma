FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1 as builder

ENV BASE_APP_DIR /workspace/go/src/github.com/kyma-project/kyma/tests/integration/dex
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./dex/ ${BASE_APP_DIR}/dex/
COPY ./licenses/ ${BASE_APP_DIR}/licenses/
COPY ./entrypoint.sh ${BASE_APP_DIR}/entrypoint.sh

#
# Build tests
#

RUN CGO_ENABLED=0 go test -c ./dex -tags=integration
RUN mkdir /tests && mv ./dex.test /tests/dex.test && mv ${BASE_APP_DIR}/licenses /tests/licenses && mv ${BASE_APP_DIR}/entrypoint.sh /tests/entrypoint.sh

FROM alpine:3.10
LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /tests

#
# Install certificates and curl
#

RUN apk update && apk add ca-certificates && apk add curl && rm -rf /var/cache/apk/*

#
# Copy binary
#

COPY --from=builder /tests /tests

#
# Run tests
#

ENTRYPOINT /tests/entrypoint.sh
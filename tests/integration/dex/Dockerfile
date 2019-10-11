FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1 as builder

ENV BASE_APP_DIR /workspace/go/src/github.com/kyma-project/kyma/tests/integration/dex
WORKDIR ${BASE_APP_DIR}

RUN apk update && apk add curl

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

RUN go test -c ./dex
RUN mkdir /tests && mv ./dex.test /tests/dex.test ${BASE_APP_DIR}/licenses /app/licenses

FROM alpine:3.8
LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /app

#
# Install certificates
#

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

#
# Copy binary
#

COPY --from=builder /tests /tests

#
# Run tests
#

ENTRYPOINT ./entrypoint.sh
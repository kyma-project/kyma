FROM golang:1.14-alpine3.12 as builder

ENV BASE_APP_DIR /workspace/go/src/github.com/kyma-project/kyma/tests/integration/dex
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY . ${BASE_APP_DIR}

#
# Build tests
#

RUN CGO_ENABLED=0 go test -c ./dex -tags=integration
RUN mkdir /tests && mv ./dex.test /tests/dex.test && mv ${BASE_APP_DIR}/licenses /tests/licenses && mv ${BASE_APP_DIR}/entrypoint.sh /tests/entrypoint.sh

FROM eu.gcr.io/kyma-project/tpi/k8s-tools:20210504-12243229
LABEL source = git@github.com:kyma-project/kyma.git

#
# Copy binary
#

COPY --from=builder --chown=nobody:nogroup /tests /tests
WORKDIR /tests

#
# Install certificates and curl
#

RUN apk update &&\
	apk add ca-certificates curl &&\
	rm -rf /var/cache/apk/* &&\
	chown -R nobody:nogroup /tests

#
# Run tests
#
USER nobody
ENTRYPOINT /tests/entrypoint.sh

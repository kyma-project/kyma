FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20200117-d3885041 as builder

ENV BASE_APP_DIR /workspace/go/src/github.com/kyma-project/kyma/components/api-gateway-migrator
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./pkg/ ${BASE_APP_DIR}/pkg/
#COPY ./licenses/ ${BASE_APP_DIR}/licenses/
COPY ./cmd/ ${BASE_APP_DIR}/cmd/

#
# Build app
#

RUN CGO_ENABLED=0 GOOS=linux go build -v -o api-gateway-migrator ${BASE_APP_DIR}/cmd/main.go
RUN mkdir /app && mv ./api-gateway-migrator /app/api-gateway-migrator
# && mv ${BASE_APP_DIR}/licenses /app/licenses

FROM eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20200309-9a88ce97
LABEL source = git@github.com:kyma-project/kyma.git
WORKDIR /app

#
# Install certificates
#

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

#
# Copy binary
#

COPY --from=builder /app /app

#
# Run app
#

CMD ["/app/api-gateway-migrator"]
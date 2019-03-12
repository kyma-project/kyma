FROM alpine:3.8

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl

COPY ./e2e-upgrade-test /usr/local/bin/e2e-upgrade-test

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["e2e-upgrade-test"]

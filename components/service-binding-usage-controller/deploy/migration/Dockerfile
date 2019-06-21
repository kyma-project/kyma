FROM eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20190325-ff66a3a

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl
RUN apk update \
 && apk add jq \
 && rm -rf /var/cache/apk/*
RUN apk add bash

COPY ./service-catalog-upgrade /root/service-catalog-upgrade

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/root/service-catalog-upgrade"]
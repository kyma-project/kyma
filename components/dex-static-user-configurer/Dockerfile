FROM eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20201006-f331d22b

RUN apk add --no-cache jq apache2-utils

COPY ./scripts/config_replace.sh /

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/bin/bash", "/config_replace.sh"]

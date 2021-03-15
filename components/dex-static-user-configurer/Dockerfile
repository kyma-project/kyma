FROM eu.gcr.io/kyma-project/incubator/k8s-tools:20210310-c03fb8b6

RUN apk add --no-cache jq apache2-utils

COPY ./scripts/config_replace.sh /

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/bin/bash", "/config_replace.sh"]

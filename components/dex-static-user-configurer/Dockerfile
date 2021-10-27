FROM eu.gcr.io/kyma-project/tpi/k8s-tools:20211022-85284bf9

RUN apk add --no-cache expat>=2.4.1-r0 util-linux>=2.37.2-r1 --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main && \
    apk add --no-cache jq apache2-utils

COPY ./scripts/config_replace.sh /

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/bin/bash", "/config_replace.sh"]

FROM eu.gcr.io/kyma-project/tpi/k8s-tools:20210407-86c769bd

RUN apk add --no-cache jq apache2-utils

COPY ./scripts/config_replace.sh /

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/bin/bash", "/config_replace.sh"]

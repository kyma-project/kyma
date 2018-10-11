FROM alpine

ENV KUBE_LATEST_VERSION="v1.11.3"

RUN apk add --no-cache jq apache2-utils bash
RUN wget -q https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -O /usr/local/bin/kubectl \
	&& chmod +x /usr/local/bin/kubectl

COPY ./scripts/config_replace.sh /

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/bin/bash", "/config_replace.sh"]
FROM alpine:3.10

LABEL source=git@github.com:kyma-project/kyma.git

ENV KUBECTL_VERSION 1.16.3

RUN apk --no-cache upgrade \
    && apk --no-cache --update add curl \
    && curl -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl && chmod +x /usr/local/bin/kubectl \
    && apk --no-cache add bash openssl

COPY . /app

RUN addgroup -S xipgroup && adduser -S xip --uid 2000 -G xipgroup -h /home/xip && chown -R xip /app
USER 2000

ENTRYPOINT [ "/app/xip-patch.sh" ]
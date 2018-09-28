FROM alpine:3.8
RUN apk --no-cache upgrade

LABEL source=git@github.com:kyma-project/kyma.git

WORKDIR /root

COPY bin/app /root

ENV ARGS=""

CMD ["sh", "-c", "/root/app $ARGS"]
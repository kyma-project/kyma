FROM alpine:3.7

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl

RUN mkdir -p /data/repository /data/repository
COPY ./reposerver /root/reposerver

ENV PORT 8080
ENV CONTENT_PATH /data/repository

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/root/reposerver"]
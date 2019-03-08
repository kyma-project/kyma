FROM alpine:3.8

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl

LABEL source = git@github.com:kyma-project/kyma.git

ADD etcd-backup /

CMD ["/etcd-backup"]

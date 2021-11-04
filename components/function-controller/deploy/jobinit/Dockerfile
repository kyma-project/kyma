FROM eu.gcr.io/kyma-project/external/golang:1.16.3-alpine as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/kyma/components/function-controller \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

RUN apk add gcc libc-dev
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.14/community libgit2-dev=1.1.0-r2

WORKDIR ${BASE_APP_DIR}

#
# copy files allowed in .dockerignore
#
COPY . ${BASE_APP_DIR}/

RUN go build -ldflags "-s -w" -a -o jobinit cmd/jobinit/main.go \
    && mkdir /app \
    && mv ./jobinit /app/jobinit

# result container
FROM alpine:3.14.2

RUN apk add --no-cache ca-certificates
RUN apk add --no-cache --update --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main openssh-client
RUN apk add --no-cache  libgit2=1.1.0-r2

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /app /app

ENTRYPOINT ["/app/jobinit"]

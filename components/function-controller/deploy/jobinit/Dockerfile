# image builder base on golang:1.21.1-alpine3.18
FROM golang@sha256:0c860c7ceba62231d0f99fb92e9d7c1577f26fea794a12c75756a8f64b146e45 as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/kyma/components/function-controller \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    LIBGIT2_VERSION=1.5.2-r0

RUN apk add --no-cache gcc libc-dev
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.18/community libgit2-dev=${LIBGIT2_VERSION}

WORKDIR ${BASE_APP_DIR}

#
# copy files allowed in .dockerignore
#
COPY . ${BASE_APP_DIR}/

RUN go build -ldflags "-s -w" -a -o jobinit cmd/jobinit/main.go \
    && mkdir /app \
    && mv ./jobinit /app/jobinit

# result container base on alpine:3.17.4
FROM europe-docker.pkg.dev/kyma-project/prod/external/alpine@sha256:e95676db9e4a4f16f6cc01a8915368f82b018cc07aba951c1bd1db586c081388
ENV LIBGIT2_VERSION=1.5.2-r0

LABEL source = git@github.com:kyma-project/kyma.git

RUN apk update --no-cache && apk upgrade --no-cache
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache --update --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main openssh-client openssl
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.18/community libgit2=${LIBGIT2_VERSION}

COPY --from=builder /app /app

ENTRYPOINT ["/app/jobinit"]

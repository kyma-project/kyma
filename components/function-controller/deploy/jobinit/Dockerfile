FROM eu.gcr.io/kyma-project/external/golang:1.16.3-alpine as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/kyma/components/function-controller \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    LIBGIT2_VERSION=1.1.0-r2

RUN apk add gcc libc-dev
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.14/community libgit2-dev=${LIBGIT2_VERSION}

WORKDIR ${BASE_APP_DIR}

#
# copy files allowed in .dockerignore
#
COPY . ${BASE_APP_DIR}/

RUN go build -ldflags "-s -w" -a -o jobinit cmd/jobinit/main.go \
    && mkdir /app \
    && mv ./jobinit /app/jobinit

# result container
FROM eu.gcr.io/kyma-project/external/alpine:3.14.2
ENV LIBGIT2_VERSION=1.1.0-r2

LABEL source = git@github.com:kyma-project/kyma.git

RUN apk add --no-cache ca-certificates
RUN apk add --no-cache --update --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main openssh-client
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.14/community libgit2=${LIBGIT2_VERSION}

COPY --from=builder /app /app

ENTRYPOINT ["/app/jobinit"]

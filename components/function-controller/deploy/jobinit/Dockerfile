# image builder base on golang:1.20.3-alpine3.17
FROM europe-docker.pkg.dev/kyma-project/prod/external/golang@sha256:48c87cd759e3342fcbc4241533337141e7d8457ec33ab9660abe0a4346c30b60 as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/kyma/components/function-controller \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    LIBGIT2_VERSION=1.4.3-r0

RUN apk add --no-cache gcc libc-dev
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.16/community libgit2-dev=${LIBGIT2_VERSION}

WORKDIR ${BASE_APP_DIR}

#
# copy files allowed in .dockerignore
#
COPY . ${BASE_APP_DIR}/

RUN go build -ldflags "-s -w" -a -o jobinit cmd/jobinit/main.go \
    && mkdir /app \
    && mv ./jobinit /app/jobinit

# result container base on alpine:3.17.3
FROM europe-docker.pkg.dev/kyma-project/prod/external/alpine@sha256:b6ca290b6b4cdcca5b3db3ffa338ee0285c11744b4a6abaa9627746ee3291d8d
ENV LIBGIT2_VERSION=1.4.3-r0

LABEL source = git@github.com:kyma-project/kyma.git

RUN apk update --no-cache && apk upgrade --no-cache
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache --update --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main openssh-client openssl
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.16/community libgit2=${LIBGIT2_VERSION}

COPY --from=builder /app /app

ENTRYPOINT ["/app/jobinit"]

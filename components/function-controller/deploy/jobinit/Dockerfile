FROM eu.gcr.io/kyma-project/external/golang:1.16.0-alpine as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/kyma/components/function-controller \
CGO_ENABLED=0 \
GOOS=linux \
GOARCH=amd64

WORKDIR ${BASE_APP_DIR}

#
# copy files whitelisted in .dockerignore
#
COPY . ${BASE_APP_DIR}/

RUN go build -ldflags "-s -w" -a -o jobinit cmd/jobinit/main.go \
&& mkdir /app \
&& mv ./jobinit /app/jobinit

# result container
FROM eu.gcr.io/kyma-project/external/alpine:3.13.2 as certs

RUN apk add --no-cache \
  ca-certificates

RUN apk --no-cache add --update openssh-client --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /app /app

ENTRYPOINT ["/app/jobinit"]

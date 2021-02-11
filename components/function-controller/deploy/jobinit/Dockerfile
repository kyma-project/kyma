FROM golang:1.15-alpine as builder

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
FROM alpine:3.13

RUN apk add --no-cache \
  openssh-client \
  ca-certificates

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /app /app

ENTRYPOINT ["/app/jobinit"]
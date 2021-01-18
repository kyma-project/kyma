ARG BASE_IMAGE_NAME
FROM eu.gcr.io/kyma-project/external/golang:1.14.8-alpine as builder

RUN apk add --no-cache git procps

RUN mkdir -p /.config/dlv && touch /.config/touched && chown -R nobody /.config
RUN CGO_ENABLED=0 go get -ldflags "-s -w -extldflags '-static'" github.com/go-delve/delve/cmd/dlv

RUN echo -e "#!/bin/sh \n \
    /dlv --accept-multiclient --api-version=2 --listen=localhost:40000 --headless --log --only-same-user=false attach 1 --" >> /debug.sh \
    && chmod +x /debug.sh

FROM $BASE_IMAGE_NAME
WORKDIR /
COPY --from=builder / /
COPY --from=builder /.config ./.config
COPY --from=builder /go/bin/dlv ./dlv
COPY --from=builder /debug.sh /debug.sh
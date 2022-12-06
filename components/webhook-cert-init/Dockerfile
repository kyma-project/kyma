FROM golang:1.19.3 AS build

WORKDIR /src/
COPY main.go go.* /src/
RUN CGO_ENABLED=0 go build -o /bin/webhook-cert-init

FROM scratch
COPY --from=build /bin/webhook-cert-init /webhook-cert-init

USER 65532:65532

ENTRYPOINT ["/webhook-cert-init"]
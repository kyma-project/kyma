FROM golang:1.19.3 AS build

WORKDIR /src/
COPY main.go go.* /src/
COPY internal/ /src/internal/
RUN CGO_ENABLED=0 go build -o /bin/webhook-ca-init

FROM scratch
COPY --from=build /bin/webhook-ca-init /bin/webhook-ca-init
ENTRYPOINT ["/bin/webhook-ca-init"]
FROM golang AS build

WORKDIR /src/
COPY main.go go.* /src/
RUN CGO_ENABLED=0 go build -o /bin/exporter

FROM alpine
COPY --from=build /bin/exporter /bin/exporter
ENTRYPOINT ["/bin/exporter"]
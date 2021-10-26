FROM golang:1.17 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY main.go main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o grafanawatcher main.go

FROM eu.gcr.io/kyma-project/external/grafana/grafana:7.3.3

WORKDIR /
COPY --from=builder /workspace/grafanawatcher .
USER 65532:65532

ENTRYPOINT ["/grafanawatcher"]
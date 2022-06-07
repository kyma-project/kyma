FROM golang:1.18 as builder

WORKDIR /connectivity-test/

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 go test -v -c -o test .

FROM scratch

COPY --from=builder /connectivity-test/test /
ENTRYPOINT [ "/test" ]
CMD []

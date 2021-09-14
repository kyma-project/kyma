FROM natsio/nats-operator:0.8.2 as builder
FROM alpine:3.13.5
COPY --from=builder /usr/local/bin/nats-operator /usr/local/bin/nats-operator
CMD ["nats-operator"]

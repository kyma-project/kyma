FROM connecteverything/nats-operator:0.7.6 as builder
FROM alpine:3.13.5
COPY --from=builder /usr/local/bin/nats-operator /usr/local/bin/nats-operator
CMD ["nats-operator"]

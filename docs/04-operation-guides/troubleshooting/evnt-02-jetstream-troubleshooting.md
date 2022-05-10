---
title: JetStream backend troubleshooting
---

## Symptom

Events were not received by the consumers.

## Remedy

1. Follow the basic troubleshooting steps as mentioned in [Eventing Troubleshooting](./evnt-01-eventing-troubleshooting.md) guide.

2. Use the [nats CLI](https://github.com/nats-io/natscli) to check if the stream was created:

    1. Port forward the Kyma Eventing NATS Service to localhost. Use port `4222`. Run:
        ```bash
        kubectl -n kyma-system port-forward svc/eventing-nats 4222
        ```
    2. Use the [nats CLI](https://github.com/nats-io/natscli) to list the streams:
       ```bash
        $ nats stream ls
        ╭────────────────────────────────────────────────────────────────────────────╮
        │                                  Streams                                   │
        ├──────┬─────────────┬─────────────────────┬──────────┬───────┬──────────────┤
        │ Name │ Description │ Created             │ Messages │ Size  │ Last Message │
        ├──────┼─────────────┼─────────────────────┼──────────┼───────┼──────────────┤
        │ sap  │             │ 2022-05-03 00:00:00 │ 0        │ 318 B │ 5.80s        │
        ╰──────┴─────────────┴─────────────────────┴──────────┴───────┴──────────────╯
        ```       

    3. If the stream exists, check the timestamp of the `Last Message` that the stream received. A recent timestamp would mean that the event was published correctly.
   
    4. Check if the consumers were created and have the expected configurations.
       ```bash
       nats consumer info
       ```
       
    To correlate the consumer to the Subscription and the specific event type, check the `description` field of the consumer.

3. Check the [JetStream grafana dashboard](https://grafana.com/grafana/dashboards/14725):

    1. Port forward the Kyma Eventing NATS Service to localhost. Use port `8081`. Run:
        ```bash
        kubectl -n kyma-system port-forward svc/monitoring-grafana 8081:80
        ```
    2. On `localhost:8081` search for `NATS JetStream` dashboard. You can find the stream and consumer metrics as well as the storage and memory consumption.
    

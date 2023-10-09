---
title: NATS JetStream backend troubleshooting
---

## Symptom

Events were not received by the consumers.

## Remedy

1. Follow the diagnostic steps as mentioned in [Eventing Troubleshooting](evnt-01-eventing-troubleshooting.md).

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

       5. If the PVC storage is fully consumed and matches the stream size as shown above, the stream can no longer receive messages. Either increase the PVC storage size or set the `MaxBytes` property which removes the old messages.

3. In case you had a [custom kube-prometheus-stack installed](https://github.com/kyma-project/examples/tree/main/prometheus), Check the [JetStream grafana dashboard](https://grafana.com/grafana/dashboards/14725):

    1. [Access Grafana](https://github.com/kyma-project/examples/tree/main/prometheus#verify-the-installation).

    2. Search for `NATS JetStream` dashboard. You can find the stream and consumer metrics as well as the storage and memory consumption.
   
    3. Also search for `JetStream Event Types Summary` and `Delivery per Subscription` dashboards to visualize and debug the phase during which the events were lost.
    

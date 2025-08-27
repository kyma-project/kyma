# NATS JetStream Backend Troubleshooting

## Symptom

Events were not received by the consumers.

## Solution

1. Follow the diagnostic steps as mentioned in [Eventing Troubleshooting](evnt-01-eventing-troubleshooting.md).

2. Use the [NATS CLI](https://github.com/nats-io/natscli) to check if the stream was created:

    1. Port forward the Kyma Eventing NATS Service to localhost. Use port `4222`. Run:

        ```bash
        kubectl -n kyma-system port-forward svc/eventing-nats 4222
        ```

    2. Use the [NATS CLI](https://github.com/nats-io/natscli) to list the streams:

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

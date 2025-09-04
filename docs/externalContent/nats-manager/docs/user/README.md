# NATS Module

This module ships the NATS Manager, which is responsible for managing the lifecycle of a [NATS JetStream](https://docs.nats.io/nats-concepts/jetstream) deployment.
It observes the state of the NATS cluster and reconciles its state according to the desired state.

NATS is an infrastructure that enables the exchange of data in form of messages. JetStream is a distributed persistence system providing more functionalities and higher qualities of service on top of 'Core NATS'.
For further information about NATS and NATS JetStream, consult the [Official NATS Documentation](https://docs.nats.io/).

Kyma Eventing can use NATS as a backend to process events and send them to subscribers.

## Documentation Overview

- [General information about the NATS Manager](./01-manager.md)
- [Details how to configure the NATS module](./02-configuration.md)

There is further documentation including more technical details aimed at possible contributors:

- [General information about the setup](../contributor/development.md)
- [Guide to the project governance](../contributor/governance.md)
- [Installation guide](../contributor/installation.md)
- [Information about the test coverage](../contributor/testing.md)
- [Troubleshooting](../contributor/troubleshooting.md)

---
title: Basic architecture
type: Architecture
---

Compass is a central place which stores applications and runtimes configurations, and then propagates that information accordingly. It also plays a crucial role in establishing a trusted connection between applications and runtimes. The basic workflow looks as follows:

![Basic architecture](./assets/architecture.svg)

1. Administrator adds runtimes and applications, and configures them using Compass.
2. Agent, a component that is integrated in every Kyma runtime, continuously fetches the actual configuration from Compass. It also sends information about runtime health checks to Compass.
3. In case an application has optional webhooks configured, Compass notifies an application about any Events that concern a given application.

Compass does not participate in any business flow. After establishing a trusted connection between an application and a runtime, they communicate directly with each other.

## Scenarios

In order to connect and group your applications and runtimes, assign them to the same scenario.
A scenario is a simple label with the **scenarios** key. If an application is not explicitly assigned to any scenario, it belongs to the `default` one. Applications and runtimes are automatically removed from the `default` scenario after you assign them to any other scenario. You can assign applications and runtimes to many scenarios. See the example:

![Scenarios](./assets/scenarios.svg)

`Application 2` belongs to the `production` and `stage` scenarios, which means that both `Runtime 1` and `Runtime 2` can consume its API or Events. Communication between components that do not belong to the same scenario is forbidden. For example, `Application 3` cannot communicate with `Runtime 1`.

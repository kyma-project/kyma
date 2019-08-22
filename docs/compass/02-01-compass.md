---
title: Basic architecture
type: Architecture
---

Compass is a central place which stores configuration about applications and runtimes, and then propagates that information to applications and runtimes. It also plays a crucial role in establishing a trusted connection between applications and runtimes. The basic workflow looks as follows:

![Basic architecture](./assets/architecture.svg)

1. Administrator adds runtimes and applications, and configures them using Compass GraphQL API.
2. Agent, a component that is integrated in every Kyma runtime, continuously fetches the actual configuration from Compass. It also sends information about runtime health checks to Compass.
3. In case an application has configured optional Webhooks, Compass notifies an Application about any events that concern given Application.

Compass does not participate in any business flow. After establishing a trusted connection between application and runtime, they communicate directly with each other.

## Scenarios

In order to connect applications and runtimes, assign them to the same scenario.
A scenario is a simple label with the **scenario** key. If an application or a runtime is not explicitly assigned to any scenario, they belong to the `default` one.
After you add a scenario, they are automatically removed from the `default` scenario and moved to the new one. You can assign applications and runtimes to many scenarios. See the example:

![Scenarios](./assets/scenarios.svg)

`Application 2` belongs to the `production` and `stage` scenarios. Thanks to that, its API or Events can be consumed by both `Runtime 1` and `Runtime 2`. Communication between components that do not belong to the same scenario is forbidden. For example, `Application 3` cannot communicate with `Runtime 1`.

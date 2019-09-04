---
title: Basic architecture
type: Architecture
---

Compass is a central place which stores applications and Runtimes configurations, and then propagates that information accordingly. It also plays a crucial role in establishing a trusted connection between applications and Runtimes. The basic workflow looks as follows:

![Basic architecture](./assets/architecture.svg)

1. Administrator adds Runtimes and applications, and configures them using Compass.
2. Agent, a component that is integrated in every Kyma Runtime, continuously fetches the actual configuration from Compass. In the future releases, Agent will also send information about Runtime health checks to Compass.
3. In case an application has optional webhooks configured, Compass notifies an application about any Events that concern a given application.

Compass does not participate in any business flow. After establishing a trusted connection between an application and a Runtime, they communicate directly with each other.

## Scenarios

In order to connect and group your applications and Runtimes, assign them to the same scenario.
A scenario is a simple label with the **scenarios** key. If an application is not explicitly assigned to any scenario, it belongs to the `default` one. The application is automatically removed from the `default` scenario after you assign it to any other scenario. Runtimes by default are not assigned to any scenario. You can assign applications and Runtimes to many scenarios. See the example:

![Scenarios](./assets/scenarios.svg)

>**CAUTION:** Currently, you can connect an application only to one Runtime at the same time.  

`Application 2` belongs to the `campaign` and `marketing` scenarios, which means that both `Runtime 1` and `Runtime 2` can consume its APIs or Events, but not at the same time. If you try to connect a Runtime to an application that is already assigned to another Runtime, the communication fails. Communication between components that do not belong to the same scenario is completely forbidden. For example, `Application 3` cannot communicate with `Runtime 1`.

---
title: Detailed workflow
type: Architecture
---

The basic flow of interactions between an Application, Runtime, and Compass is as follows:

1. Connecting Application
2. Registering Runtime
3. Changing configuration

## Connecting Application

Connecting an Application consists of two phases: Application pairing and API registration. In the process of connecting a new Application, two Compass components are involved: the Director and the Connector.

### Application pairing phase

Application pairing phase is a process of creating a new Application and establishing a trusted connection between the Application and Compass. The workflow looks as follows:

![](./assets/app-pairing.svg)

1. Administrator sends a request to register a new Application in Compass.
2. Director registers the Application.
3. Director sends back Application details, along with its unique ID.
4. Administrator requests Application pairing with the Connector.
5. Connector generates a one-time token and sends it back to the Administrator.
6. Administrator passes the one-time token to the Application.
7. Application uses this token to establish a trusted connection with Compass.

### API registration phase

API registration phase is a process of registering new API and Event Definitions, which consists of two steps:

![](./assets/api-registration.svg)

1. Application sends a request to the Director to register an API or Event Definition.
2. Director returns the operation result to the Application.

## Registering Runtime

The process of registering a new Runtime looks as follows:

![](./assets/runtime-creation.svg)

1. Administrator sends a request to provision a new Runtime.
2. Runtime Provisioner requests Runtime configuration from the Director.
3. Director returns Runtime configuration to the Runtime Provisioner.
4. Runtime Provisioner requests a one-time token from the Connector.
5. Connector generates a one-time token for the Runtime.
6. Connector returns the token to the Runtime Provisioner.
7. Runtime Provisioner provisions the Runtime.
8. Runtime Provisioner injects Runtime configuration along with the one-time token to the Runtime Agent.
9. Runtime Agent uses the token to set up a trusted connection between Compass and the Runtime.

When the Runtime is ready, Runtime Agent notifies the Director about the Runtime status. When the Director receives a notification that a Runtime is ready, it passes the notification to every Application in a group assigned to this Runtime using Application Webhook API.

## Changing configuration

The following section describes how configuration updates work for Applications and Runtimes.

### Application configuration update

There are two options for updating Application configuration:
- Periodically fetching Application configuration
- Exposing Application Webhook API to get notifications on configuration updates

In the first case, the Application periodically pulls configuration details, such as **eventURL**, for connected Runtimes.

![](./assets/app-configuration-update.svg)

In the second case, if configuration for any connected Runtime changes, Application Webhook API notifies an Application that new configuration details are available. The following diagram shows the interaction between the Runtime Agent, the Director, and an Application when a new Runtime is configured successfully:

![](./assets/runtime-notification.svg)

1. Runtime Agent sends a request to change Application configuration.
2. Director notifies the Application about the request.
3. Application changes configuration for the Runtime.
4. Runtime Agent gets a new configuration from the Director.

### Runtime configuration update

Runtime Agent gets Runtime configuration details from the Director and applies the configuration asynchronously. The configuration details include information such as the list of Applications and their credentials. Runtime Agent periodically checks for new configuration details and applies the changes.

![](./assets/runtime-configuration-update.svg)

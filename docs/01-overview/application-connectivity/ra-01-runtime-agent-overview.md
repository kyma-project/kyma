---
title: Runtime Agent
---

Runtime Agent is a Kyma component that connects to [Compass](https://github.com/kyma-incubator/compass). It is an integral part of every Kyma Runtime in the [Compass mode](README.md) and it fetches the latest configuration from Compass. It also provides Runtime-specific information that is displayed in the Compass UI, such as Runtime UI URL, and it provides Compass with Runtime configuration, such as Event Gateway URL, that should be passed to an Application. To learn more, read the section on [configuring the Runtime](../../05-technical-reference/ra-01-configuring-runtime.md).

The main responsibilities of the component are:
- Establishing a trusted connection between the Kyma Runtime and Compass
- Renewing a trusted connection between the Kyma Runtime and Compass
- Synchronizing with the [Director](https://github.com/kyma-incubator/compass/blob/master/docs/compass/02-01-components.md#director) by fetching new Applications from the Director and creating them in the Runtime, and removing from the Runtime Applications that no longer exist in the Director.

## Useful links

If you're interested in learning more about Runtime Agent, follow these links to:

- Perform some simple and more advanced tasks:

    - [Enable Kyma with Runtime Agent](../../04-operation-guides/operations/ra-01-enable-kyma-with-runtime-agent.md)
    - [Establish a secure connection with Compass](../../03-tutorials/00-application-connectivity/ra-01-establish-secure-connection-with-compass.md)
    - [Maintain a secure connection with Compass](../../03-tutorials/00-application-connectivity/ra-02-maintain-secure-connection-with-compass.md)
    - [Revoke a client certificate (RA)](../../03-tutorials/00-application-connectivity/ra-03-revoke-client-certificate.md)
    - [Configure Runtime Agent with Compass](../../03-tutorials/00-application-connectivity/ra-04-configure-runtime-agent-with-compass.md)
    - [Reconnect Runtime Agent with Compass](../../03-tutorials/00-application-connectivity/ra-05-reconnect-runtime-agent-with-compass.md)
    
- Analyze Runtime Agent specification and configuration files:

    - [Compass Connection](../../05-technical-reference/00-custom-resources/ra-01-compassconnection.md) custom resource (CR)
    - [Connection with Compass](../../05-technical-reference/00-configuration-parameters/ra-01-connection-with-compass.md) 

- Understand technicalities behind the Runtime Agent implementation:

    - [Runtime Agent workflow](../../05-technical-reference/00-architecture/ra-01-runtime-agent-workflow.md)
    - [Configuring the Runtime](../../05-technical-reference/ra-01-configuring-runtime.md)

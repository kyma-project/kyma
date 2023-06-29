---
title: Runtime Agent
---

Runtime Agent is a Kyma component that connects to [Compass](https://github.com/kyma-incubator/compass). It is an integral part of every Kyma Runtime in the [Compass mode](README.md) and it fetches the latest configuration from Compass. It also provides Runtime-specific information that is displayed in the Compass UI, such as Runtime UI URL, and it provides Compass with Runtime configuration, such as Event Gateway URL, that should be passed to an Application. To learn more, read the section on [configuring the Runtime](../../05-technical-reference/ra-01-configuring-runtime.md).

The main responsibilities of the component are:
- Establishing a trusted connection between the Kyma Runtime and Compass
- Renewing a trusted connection between the Kyma Runtime and Compass
- Synchronizing with the [Director](https://github.com/kyma-incubator/compass/blob/master/docs/compass/02-01-components.md#director) by fetching new Applications from the Director and creating them in the Runtime, and removing from the Runtime Applications that no longer exist in the Director.
---
title: Remote Environments
type: Architecture
---

An external system is connected to Kyma using a dedicated Remote Environment. Each system has one dedicated Remote Environment.

The Remote Environment is handling aspects of the security and integration with other Kyma components, like Service Catalog and Event bus.

The Remote Environment is creating a coherent entity from underlying Application Connector's components for a connected system and ensures a separation between different environments.

All Remote Environments are stored using Kubernetes Custom Resource Definition (CRD). You can find more details here [Remote Environment CRD](040-remote-environment-custom-resource.md)

The Remote Environment can be mapped to many Kyma Environments using the [Environment Mapping CRD](041-cr-environment-mapping.md).


---
title: Container registries
---

By default, Serverless uses PersistentVolume (PV) as the internal registry to store Docker images for Functions. The default storage size of a single volume is 20 GB. This internal registry is suitable for local development.

If you use Serverless for production purposes, it is recommended that you use an external registry, such as Docker Hub, Google Container Registry (GCR), or Azure Container Registry (ACR).

Serverless supports two ways of connecting to an external registry:

- [You can set up an external registry before installation](../../../03-tutorials/serverless/svls-07-set-external-registry.md).

  In this scenario, you can use Kyma overrides to change the default values supplied by the installation mechanism.

- [You can switch to an external registry at runtime](../../../03-tutorials/serverless/svls-08-switch-to-external-registry.md).

  In this scenario, you can change the registry on the fly, with Kyma already installed on your cluster. This option gives you way more flexibility and control over the choice of an external registry for your Functions.

>**TIP:** For details, read about [switching registries at runtime](../../../05-technical-reference/svls-03-switching-registries.md).

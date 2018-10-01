---
title: Kyma and Knative - brothers in arms
type: Overview
---

Integration with Knative is a step towards Kyma modularization and the "slimming" approach which aims to extract some out-of-the-box components and provide you with a more flexible choice of tools to use in Kyma.

### Background

Both Kyma and Knative are Kubernetes and Istio-based systems that offer development and eventing platforms. The main difference, however, is their focus. While Knative concentrates more on providing the building blocks for running serverless workloads, Kyma focuses on integrating those blocks with external services and applications.

The diagram shows dependencies between the components:

![kyma-knative](./assets/kyma-knative.svg)

### Planned changes

The nearest plan for Kyma and Knative cooperation is to replace Kubeless with the Knative technology. The plan assumes to take the Knative `Serving` and `Build` components as the building blocks for the new architecture, and to introduce new custom components to bridge any existing gaps.

These new components include:
- a custom build template that provides the function interface available in Kubeless
- a custom Docker registry to store the build artifacts
- a storage solution to store the function code (a Git repository or a blob storage like Minio or S3)

The implementation process involves multiple components and will follow these steps:
1. Integration of Knative `Serving` and `Build` components as optional Kyma modules.
2. Enabling other parts like Docker registry, blob storage, and build template (the order is not defined yet).
3. Creation of a new, forked lambda UI, adjusted to Knative requirements.
4. Removal of Kubeless when the integration completes successfully and no Kyma components use it anymore.

Other planned changes concerning Kyma and Knative cooperation involve providing configuration options to allow Istio deployed with Knative to work on Kyma, and extracting Kyma eventing to fully integrate it with Knative eventing. The eventing integration will provide more flexibility on deciding which messaging implementation to use (NATS, Kafka, or any other).

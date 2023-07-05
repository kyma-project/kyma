# Serverless naming convention

## Problem overview

Currently, Serverless in Kyma consists of two projects:

- [function-controller](https://github.com/kyma-project/kyma/tree/main/components/function-controller) -
  responsible for running a Function on a Kubernetes cluster
- [serverless-manager](https://github.com/kyma-project/serverless-manager) - responsible for installation and
  configuration of Serverless

Additionally, we have 3rd-party additional components like [KEDA](https://keda.sh/)

In Serverless, we overuse the word "controller" which causes confusion and requires clarification. Saying "controller", we refer to:

- Serverless reconcile loop
- Serverless Pod with the reconcile loop
- the Function Controller component in the `kyma` directory

## Goal

The goal of this document is to clarify the naming convention in Serverless and define its elements to avoid confusion and make it more logical.

## Proposal

The proposed naming conventions refer to different architecture layers of the whole project. See the [architecture](./assets/kubebuilder-architecture.png) diagram for details.

### Project naming convention

This section refers to the high-level architecture elements, namely to the main projects:

- Serverless - the new naming convention for the `function-controller`. Serverless is responsible for running a Function on a Kubernetes cluster. It can contain its own
  CRD.
- Serverless-operator - the new naming convention for `serverless-manager`. Serverless-operator installs and configures Serverless.
- Kyma-Keda-operator - the operator which installs and configures [KEDA](https://keda.sh/).

### Component naming convention

This section refers to the Serverless components:

- Controller - responsible for creating and configuring k8s resources to finally run a function on a cluster. It is responsible for the reconciliation of the Function CR.
- Webhook - responsible for defaulting, validation, and conversion of the Function CR, mutating the external registry Secret, and reconciling certificates.

Proposed naming convention:

Deployment with the controller in charge:

- ${component_name}-controller

In the case of introducing a separate CRD and a separate deployment:

- ${component_name}-{crd_name}-controller

Deployment with the webhook as main responsibility:

- ${component_name}-webhook

Deployment with both controller and webhook:

- ${component_name}

> **NOTE:** I decided to go with the pure component name as it contains both the controller and webhook responsibilities and from the technical perspective it's very similar to the component itself. It might be confusing, and I am open to other proposals.

### Kubebuilder component naming convention

Serverless uses Kubebuilder to build a controller and/or webhook. This section describes the naming convention of the most detailed project layer, namely Kubebuilder components.

Looking at the [architecture](./assets/kubebuilder-architecture.png) diagram, you can see that a program consists of a **process** which includes a **manager**.
The **manager** can include 2 components:

- Controller, which focuses on the reconciliation of a given Kubernetes resource. It uses predicates and the reconciler.
- Webhook, which works with `AdmissionRequests`.

Proposed naming convention:

For the controller reconcile loop inside the manager:

- ${crd_name}-reconcile
- ${component_name}-${crd_name}-reconcile

For the webhook inside the manager:

- ${component_name}-validaton-webhook
- ${component_name}-${crd_name}-validaton-webhook
- ${crd_name}-validaton-webhook

For serverless runtimes:

- ${runtime_name}, eg.: `python310`

## Summary

The table lists the terms from the most general to the most detailed ones:

| component name                | responsibility                                 |
|-------------------------------|------------------------------------------------|
| serverless                    | the product, such as Keda                      |
| serverless-operator           | serverless installer                           |
| serverless-controller         | serverless main reconciliation loop deployment |
| serverless-webhook            | serverless webhook deployment                  |
| serverless-reconciler         | serverless reconciliation loop                 |
| serverless-validation-webhook | serverless validation webhook                  |
| serverless-defaulting-webhook | serverless defaulting webhook                  |
| serverless-conversion-webhook | serverless conversion webhook                  |
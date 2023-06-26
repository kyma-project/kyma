# Serverless naming proposal

## Problem

Serverless consist of two projects:

- [function-controller](https://github.com/kyma-project/kyma/tree/main/components/function-controller), which is
  responsible for running function on k8s cluster
- [serverless-manager](https://github.com/kyma-project/serverless-manager), which is responsible for installation and
  configuration of serverless
- additionally we have 3rd-party additional components like [keda](https://keda.sh/)

Example of confusion with name `controller`. In our case we name `controller`:

- serverless reconcile loop
- serverless pod with reconcile loop
- component in kyma/directory

## Goal

The target of this document is the attempt to unify naming convention in serverless component to avoid confusion and
make it more logical.

# Proposal

## Components naming convention

The proposed naming conversion is:

- Serverless - component which is doing some buisness logic related to functions on k8s cluster and can contain its own
  CRD.
- Serverless-operator - name of the component which install and configure Serverless.
- Kyma-Keda-operator - the operator install and configure [keda](https://keda.sh/)

## Deployment naming description

Serverless consist of two deployment:

- controller - responsible for creating, configuring k8s resources to finally run function on a cluster. The main
  responsibility is reconciliation of functions cr.
- defaulting/validation/conversion webhooks responsible for defaulting/validation for function CR, mutating external
  registry secret but also reconcile certificates. The main responsibility is doing webhook things.

Serverless uses kubebuilder to build the program and has the
following [architecture](./assets/kubebuilder-architecture.png).

The program consists of `process` which have `manager`.
The `manager` can have 2 components:

- Controller, which focuses on reconciliation of given k8s resource. Uses predicates and reconciler.
- Webhook, which works with AdmissionRequests

## Components naming convention

My proposal about naming:

Deployment with controller as main responsibility:

- ${component_name}-controller

In case of introducing separate CRD and separate deployment:

- ${component_name}-{crd_name}-controller

Deployment with webhook as main responsibility:

- ${component_name}-webhook

Deployment with both of those components:

- ${component_name}

I decided to go with pure component name as it contains two responsibilities and from technical perspective it's very
similar to the component.
It might be confusing, and I am open to hear some proposals.

For controller reconcile loop inside the manger:

- ${crd_name}-reconcile
- ${component_name}-${crd_name}-reconcile

For webhook inside the manager:

- ${component_name}-validaton-webhook
- ${component_name}-${crd_name}-validaton-webhook
- ${crd_name}-validaton-webhook

For serverless runtimes:

- ${runtime_name}, eg.: `python310`

## Summary

Table is arrange from the least technical to the most technical terms.

| component name                | responsibility                                 |
|-------------------------------|------------------------------------------------|
| serverless                    | The product like `Keda`                        |
| serverless-operator           | serverless isntaller                           |
| serverless-controller         | serverless main reconciliation loop deployment |
| serverless-webhook            | serverless webhook deployment                  |
| serverless-reconciler         | serverless reconciliation loop                 |
| serverless-validation-webhook | serverless validation webhook                  | 
| serverless-defaulting-webhook | serverless defaulting webhook                  | 
| serverless-conversion-webhook | serverless conversion webhook                  | 

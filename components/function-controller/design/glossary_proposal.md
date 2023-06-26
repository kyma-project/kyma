# Serverless dictionary

The target of this document is the attempt to unify naming convention in serverless component to avoid confusion.
Example of confusion with name `controller`. In our case we name `controller`:

- serverless reconcile loop
- serverless pod with reconcile loop
- component in kyma/directory

Serverless consist of two projects:

- serverless itself, which is responsible for running function on k8s cluster
- serverless-manager, which is responsible for instalation and configuration of serverless
- additionally we have 3rd-party additional components like keda

## Components naming convention

The proposed naming conversion is:

- Serverless - component which is doing some buisness logic related to functions on k8s cluster and can contain its own
  CRD.
- Serverless-operator - name of the component which install and configure
  Serverless. *[Operator framework](https://operatorframework.io/) *The Operator Framework is an open source toolkit to
  manage Kubernetes native applications, called Operators, in an effective, automated, and scalable way.
- Kyma-Keda-operator - the operator install and configure [keda](https://keda.sh/)

## Services naming description

Serverless consist of two services:

- controller - responsible for creating, configuring k8s resources to finally run function on a cluster. The main
  responsibility is controlling of functions cr
- defaulting/validation/conversion webhooks responsible for defaulting/validation for function CR, mutating external
  registry secret but also reconcile certificates. The main responsibility is doing webhook things.

Serverkess uses kubebuilder to build the services and has the
following [architecture](./assets/kubebuilder-architecture.png).

The service consists of `process` which have `manager`.
The `manager` can have 2 components:

- Controller, which focuses on reconciliation of given k8s resoruce. Uses predicate and reconciler.
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

- ${component_name} TODO: czy to jest dobra nazwa, przykład wardena? Wprowadza się tutaj podwójna nazwę, jedną dla
  technicznego komponentu a druga dla produktu. Może to jednak jest dobry pomysł?

For controller reconcile loop inside the manger:

- ${crd_name}-reconcile
- ${component_name}-${crd_name}-reconcile

For webhook inside the manager:

- ${component_name}-validaton-webhook
- ${component_name}-${crd_name}-validaton-webhook
- ${crd_name}-validaton-webhook

Summary, from the top to the bottom:

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

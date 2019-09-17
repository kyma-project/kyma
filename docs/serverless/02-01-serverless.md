---
title: Architecture
---

The following diagram illustrates a generic serverless implementation.

![General serverless architecture](./assets/serverless_general.svg)

The application flow takes place on the client side. Third parties handle the infrastructural logic. Custom logic can process updates and encapsulate databases. Authentication is an example of custom logic. Third parties can also handle business logic. A hosted database contains read-only data that the client reads. None of this functionality runs on a single, central server. Instead, the client relies on FaaS as its resource.

The following diagram shows an example of tasks that lambdas can perform in Kyma after a user invokes them.

![Lambdas in Kyma](./assets/lambda_example.svg)

First, the user invokes the exposed lambda endpoint. Then, the lambda function can carry out a number of tasks, such as:

* Retrieving cart information from Enterprise Commerce
* Retrieving stock details
* Updating a database

## Open source components

Kyma is comprised of several open source technologies to provide extensive functionality.

### Kubeless

Kubeless is the serverless framework integrated into Kyma that allows you to deploy lambda functions. These functions run in Pods inside the Kubeless controller on a node, which can be a virtual or hardware machine.

Kubeless also has a command line interface. Use Node.js to create lambda functions.

### Istio

Istio is a third-party component that makes it possible to expose and consume services in Kyma. See the [Istio documentation](https://istio.io) to learn more. Istio helps create a network of deployed services, called a service mesh.

In Kyma, functions run in Pods. Istio provides a proxy for specified pods that talk to a pilot. The pilot confirms whether access to the pod is permissible as per the request. In the diagram, Pod B requests access to Pod A. Pod A has an Istio proxy that contains a set of instructions on which services can access Pod A. The Istio proxy also notifies Pod A as to whether Pod B is a part of the service mesh. The Istio Proxy gets all of its information from the Pilot.

![Istio architecture](./assets/istio.svg)

### NATS

The Event Bus in Kyma monitors business events and trigger functions based on those events. At the heart of the Event Bus is NATS, an open source, stand-alone messaging system. To learn more about NATS, visit the [NATS website](https://nats.io).

The following diagram demonstrates the Event Bus architecture.

![Event Bus architecture](./assets/nats.svg)

The Event Bus exposes an HTTP endpoint that the system can consume. An external event, such as a subscription, triggers the Event Bus. A lambda function works with a push notification, and the subscription handling of the Event Bus processes the notification.

### Knative

Knative is a platform which manages serverless workloads inside Kubernetes environments. Due to its high versatility, Knative can manage the whole lifecycle of a lambda function, from building the source code to serving it.

Kyma bridges the gap between developers and Knative using its own [Function Controller](https://github.com/kyma-project/kyma/blob/master/components/function-controller/README.md) which uses the Function custom resource to facilitate the deployment of lambda functions.

For details, see the [Knative project](https://knative.dev/docs/) documentation.

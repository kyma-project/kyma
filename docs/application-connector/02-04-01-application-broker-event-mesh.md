---
title: Application Broker
type: Architecture
---

## Provisioning and binding for an event ServiceClass

This ServiceClass has a **bindable** parameter set to `false` which means that after provisioning a ServiceClass in the Namespace, given events are ready to use for all services. The provisioning workflow for an event ServiceClass consists of the following steps:

1. Select a given event ServiceClass from the Service Catalog.
2. Provision this ServiceClass by creating a ServiceInstance in the given Namespace.
3. The Service Catalog sends a provisioning request to the Application Broker with the Application and the Namespace details.
4. The Application Broker labels the Namespace with `knative-eventing-injection=enabled`, which will trigger the installation of the default Knative Eventing Broker in that Namespace.
5. The Application Broker creates a Knative Subscription in the `kyma-integration` Namespace to wire the Application's HTTPSource and the Namespace's default Knative Eventing Broker.
6. Create a Knative Trigger from the Kyma Console UI for a Lambda with a particular event type.
7. The Application sends events to the Application Connector.
8. The Application Connector sends events to Application's HTTPSource deployed inside the `kyma-integration` Namespace.
9. The Application's HTTPSource sends events the Applications Knative Channel inside the `kyma-integration` Namespace.
10. The Knative Channel sends events to the user Namespace's default Knative Eventing Broker. This happens as a result of the created Knative Subscription by the Application Broker.
11. The Knative Trigger delivers events to the Lambda for a particular event type.

![Event Service Class](./assets/007-AB-event-mesh.svg)

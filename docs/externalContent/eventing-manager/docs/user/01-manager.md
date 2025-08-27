# Eventing Manager

This module ships Eventing Manager. Once this module is added, it provides the functionality to send and receive events.

## Module Lifecycle

Upon starting Eventing Manager, the controller (following the [Kubebuilder concept](https://book.kubebuilder.io/architecture.html)) creates, watches, and reconciles the following resources:

- ConfigMap
- Secret
- Service
- Stateful Set
- Destination Rule
- Pod Disruption Budget

Eventing Manager reacts to changes in the Eventing custom resource (CR) to adapt the resources mentioned above to the desired state. For details on how to configure Eventing using the CR, visit the [Configuration documentation](02-configuration.md).

![Eventing Manager reconciliation loop](../assets/reconcileLoop.svg)

## Backend

Eventing Manager supports switching between two different backends. [NATS](https://nats.io/about/), an open source messaging system, and [EventMesh](https://help.sap.com/docs/event-mesh/event-mesh/what-is-sap-event-mesh), a SAP solution for event-based architecture.

- **NATS**
  
  If you want to use Eventing Manager with NATS backend, you must deploy the [NATS module](https://github.com/kyma-project/nats-manager).

- **SAP EventMesh**

  If you want to use Eventing Manager with EventMesh backend, you must deploy the [API-Gateway module](https://github.com/kyma-project/api-gateway).

For more information about the possible configuration of Eventing Manager using NATS or EventMesh backend, refer to the [backend configuration](02-configuration.md#reference).

## Removing the Module

The module cannot be removed as long as Subscription CRs exist. After the user cleans up all the subscriptions, the Eventing module can be removed. The module takes care of cleaning up all resources owned by it.

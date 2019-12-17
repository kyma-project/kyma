---
title: Knative Eventing Mesh
type: Details
---

## Overview

The use of Knative Eventing shifts the eventing paradigm in Kyma to an eventing mesh with isolated fault domains, access control and dynamic event routing where many senders inject events into the mesh from multiple source points, and many subscribers receive all or a subset of those events based on filters and access control. 
Thanks to the concepts of [Knative Broker and Trigger](https://knative.dev/docs/eventing/broker-trigger/) the process of Event publishing and consumption runs smoother and greatly improves the overall performance.  



## Architecture

The diagram shows you the Event flow along with the underlying structure which allows Event consumption and publishing.  

![Event Service Class](./assets/knative-event-mesh.svg)

The provisioning workflow for an Event ServiceClass consists of the following steps:

1. The user selects an Event ServiceClass from the Service Catalog. 
2. The user provisions this ServiceClass by creating a ServiceInstance in the Namespace. The ServiceClass has a **bindable** parameter set to `false` which means that after [provisioning a ServiceClass in the Namespace](/components/service-catalog/#details-provisioning-and-binding), the Events are ready to use for all services.
3. The Service Catalog sends a provisioning request with the Application and Namespace details to the Application Broker.
4. The Application Broker labels the Namespace with `knative-eventing-injection=enabled`, which triggers the installation of the default Knative Eventing Broker in that Namespace.
5. The Application Broker creates a Knative Subscription in the `kyma-integration` Namespace to wire the Application's [HTTP Adapter Source](https://github.com/kyma-project/kyma/tree/master/components/event-sources/adapter/http) and the Namespace's default Knative Eventing Broker.
6. Using the Kyma Console, the user creates a Knative Trigger for a Lambda with a particular event type.
7. As soon as the Application CR is created, the Application Operator creates a HTTP event source. This event source exposes an HTTP endpoint that receives Cloud Events and forwards them to the Knative Event Mesh.


Here is how the Events are processed: 

1. The Application sends Events to the Application Connector.
2. The Application Connector forwards events to Application's HTTP Adapter Source deployed inside the `kyma-integration` Namespace.
3. The Application's HTTP Adapter Source sends events the Applications Knative Channel inside the `kyma-integration` Namespace.
4. The Knative Channel sends events to the user Namespace's default Knative Eventing Broker. This happens as a result of the created Knative Subscription by the Application Broker.
5. The Knative Trigger delivers events to the lambda function.
---
title: Basic concepts
type: Details
---

The following resources are involved in Event transfer and validation in Kyma:

* **EventActivation** is a custom resource controller that the Application Broker (AB) creates. Its purpose is to define Event availability in a given Namespace.

* **NATS Streaming** is an open source, log-based streaming system that serves as a database allowing the Event Bus to store and transfer the Events on a large scale.

* **Persistence** is a back-end storage volume for NATS Streaming that stores Events. When the Event flow fails, the Event Bus can resume the process using the Events saved in Persistence.

* **Publish** is an internal Event Bus service that transfers the enriched Event from a given external solution to NATS Streaming.

* **Push** is an application responsible for receiving Events from NATS Streaming in the Event Bus. Additionally, it delivers the validated Events to the lambda or the service, following the trigger from the Subscription custom resource. The Events are delivered to the lambda or the service through the Envoy proxy sidecar with mTLS enabled.

* **Subscription** is a custom resource that the lambda or service creator defines to subscribe a given lambda or a service to particular types of Events.

* **Sub-validator** is a Kubernetes deployment. It updates the status of the Subscription custom resource with the EventActivation status. Depending on the status, `push` starts or stops delivering Events to the lambda or the service webhook.

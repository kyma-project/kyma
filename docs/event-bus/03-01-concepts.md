---
title: Basic concepts
type: Details
---

The following resources are involved in Event transfer and validation in Kyma:

* **EventActivation** is a custom resource controller that the Application Broker (AB) creates. Its purpose is to define Event availability in a given Namespace.

* **NATS Streaming** is an open source, log-based streaming system that serves as a database allowing the Event Bus to store and transfer the Events on a large scale.

* **Persistence** is a backend storage volume for NATS Streaming that stores Events. When the Event flow fails, the Event Bus can resume the process using the Events saved in Persistence.

* **Publish** is an internal Event Bus service that transfers the enriched Event from a given external solution to NATS Streaming.

* **Subscription** is a custom resource that the lambda or service creator defines to subscribe a given lambda or a service to particular types of Events.

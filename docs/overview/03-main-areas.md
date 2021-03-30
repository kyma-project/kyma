---
title: Main areas
type: Overview
---

Kyma is built of numerous components but these three drive it forward:

### Application connectivity

    - Simplifies and secures the connection between external systems and Kyma
    - Registers external Events and APIs in the Service Catalog and simplifies the API usage
    - Provides asynchronous communication with services and Functions deployed in Kyma through Events
    - Manages secure access to external systems
    - Provides monitoring and tracing capabilities to facilitate operational aspects

  ![connectivity](./assets/app-connectivity.svg)

### Service consumption

    - Connects services from external sources
    - Unifies the consumption of internal and external services thanks to compliance with the Open Service Broker standard
    - Provides a standardized approach to managing the API consumption and access
    - Eases the development effort by providing a catalog of API and Event documentation to support automatic client code generation

### Serverless

    - Ensures quick deployments following a Function approach
    - Enables scaling independent of the core applications
    - Gives a possibility to revert changes without causing production system downtime
    - Supports the complete asynchronous programming model
    - Offers loose coupling of Event providers and consumers
    - Enables flexible application scalability and availability

The Serverless component allows you to reduce the implementation and operation effort of an application to the absolute minimum. Kyma Serverless provides a platform to run lightweight Functions in a cost-efficient and scalable way using JavaScript and Node.js. Serverless in Kyma relies on Kubernetes resources like [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/), [Services](https://kubernetes.io/docs/concepts/services-networking/service/) and [HorizontalPodAutoscalers](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) for deploying and managing Functions and [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) for creating Docker images.

  ![serverless](./assets/serverless.svg)

### Eventing

Eventing allows you to easily integrate external applications with Kyma. Under the hood, it implements [NATS](https://docs.nats.io/) to ensure Kyma receives business events from external sources and is able to trigger business flows using Functions or services.

  ![eventing](./assets/eventing.svg)

## See also

- Kyma components

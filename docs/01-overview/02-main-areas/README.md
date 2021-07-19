---
title: Main areas
---

Kyma consists of these main areas and components:

![areas](./assets/kyma-areas.svg)

### What is Application Connectivity?

- Simplifies and secures the connection between external systems and Kyma
- Registers external events and APIs and simplifies the API usage
- Provides asynchronous communication with services and Functions deployed in Kyma through events
- Manages secure access to external systems
- Provides monitoring and tracing capabilities to facilitate operational aspects

### What is API exposure?

The API exposure in Kyma is based on the API Gateway component that aims to provide a set of functionalities which allow developers to expose, secure, and manage their APIs in an easy way. The main element of the API Gateway is the API Gateway Controller which exposes services in Kyma.

### What is Eventing?

Eventing allows you to easily integrate external applications with Kyma. Under the hood, it implements [NATS](https://docs.nats.io/) to ensure Kyma receives business events from external sources and is able to trigger business flows using Functions or services.

### What is Observability?

Kyma comes with tools that give you the most accurate and up-to-date monitoring, logging and tracing data.

- [Prometheus](https://prometheus.io/) is the open source monitoring and alerting toolkit that provides the telemetry data. This data is consumed by different add-ons, including [Grafana](https://grafana.com/) for analytics and monitoring, and [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) for handling alerts.
- For logging, Kyma uses [Loki](https://github.com/grafana/loki), a Prometheus-like log management system.
- With the [Jaeger](https://github.com/jaegertracing) distributed tracing system, you can analyze the path of a request chain going through your distributed applications. This information helps you, for example, to troubleshoot your applications, or optimize the latency and performance of your solution.

For more details, see [Observability](./observability/obsv-01-telemetry-observability-overview.md).

### What is Serverless?

- Ensures quick deployments following a Function approach
- Enables scaling independent of the core applications
- Gives a possibility to revert changes without causing production system downtime
- Supports the complete asynchronous programming model
- Offers loose coupling of Event providers and consumers
- Enables flexible application scalability and availability

Serverless in Kyma allows you to reduce the implementation and operation effort of an application to the absolute minimum. It provides a platform to run lightweight Functions in a cost-efficient and scalable way using JavaScript and Node.js. Serverless in Kyma relies on Kubernetes resources like [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/), [Services](https://kubernetes.io/docs/concepts/services-networking/service/) and [HorizontalPodAutoscalers](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) for deploying and managing Functions and [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) for creating Docker images.

### What is Service Consumption?

- Connects services from external sources
- Unifies the consumption of internal and external services thanks to compliance with the Open Service Broker standard
- Provides a standardized approach to managing the API consumption and access
- Eases the development effort by providing a catalog of API and event documentation to support automatic client code generation

### What is Service Mesh?

The Service Mesh is an infrastructure layer that handles service-to-service communication, proxying, service discovery, traceability, and security, independently of the code of the services. Kyma uses the [Istio](https://istio.io/) Service Mesh that is customized for the specific needs of the implementation.

### ...And what about the UI?

Kyma provides two interfaces that you can use for interactions:

- [Console UI](link) - a web-based administrative UI for the Kyma functionality and to manage the basic Kubernetes resources.
- [Kyma CLI](link) - a CLI to execute various Kyma tasks, such as installing or upgrading Kyma.

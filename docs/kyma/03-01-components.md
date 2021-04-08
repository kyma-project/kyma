---
title: Components
type: Details
---

Kyma is built on the foundation of the best and most advanced open-source projects which make up the components readily available for customers to use.
This section describes the Kyma components.

## Security

Kyma security enforces role-based access control (RBAC) in the cluster. [Dex](https://github.com/dexidp/dex) handles the identity management and identity provider integration. It allows you to integrate any [OpenID Connect](https://openid.net/connect/) or SAML2-compliant identity provider with Kyma using [connectors](https://github.com/dexidp/dex#connectors). Additionally, Dex provides a static user store which gives you more flexibility when managing access to your cluster.

## Service Catalog

The Service Catalog lists all of the services available to Kyma users through the registered [Service Brokers](/components/service-catalog/#overview-service-brokers). Use the Service Catalog to provision new services in the
Kyma [Kubernetes](https://kubernetes.io/) cluster and create bindings between the provisioned service and an application.

## Helm Broker

The Helm Broker is a Service Broker which runs in the Kyma cluster and deploys Kubernetes native resources using [Helm](https://github.com/kubernetes/helm) and Kyma addons. An addon is an abstraction layer over a Helm chart which allows you to represent it as a ClusterServiceClass in the Service Catalog. Use addons to install GCP, Azure
and AWS Service Brokers in Kyma.

## Application Connector

The Application Connector is a proprietary Kyma solution. This endpoint is the Kyma side of the connection between Kyma and the external solutions. The Application Connector allows you to register the APIs and the Event Catalog, which lists all of the available events, of the connected solution. Additionally, the Application Connector proxies the calls from Kyma to external APIs in a secure way.

## Eventing

Eventing allows you to easily integrate external applications with Kyma. Under the hood, it implements [NATS](https://docs.nats.io/) to ensure Kyma receives business events from external sources and is able to trigger business flows using Functions or services.

## Service Mesh

The Service Mesh is an infrastructure layer that handles service-to-service communication, proxying, service discovery, traceability, and security, independently of the code of the services. Kyma uses the [Istio](https://istio.io/) Service Mesh that is customized for the specific needs of the implementation.

## Serverless

The Serverless component allows you to reduce the implementation and operation effort of an application to the absolute minimum. Kyma Serverless provides a platform to run lightweight Functions in a cost-efficient and scalable way using JavaScript and Node.js. Serverless in Kyma relies on Kubernetes resources like [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/), [Services](https://kubernetes.io/docs/concepts/services-networking/service/) and [HorizontalPodAutoscalers](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) for deploying and managing Functions and [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) for creating Docker images.

## Monitoring

Kyma comes bundled with tools that give you the most accurate and up-to-date monitoring data. [Prometheus](https://prometheus.io/) open source monitoring and alerting toolkit provides this data, which is consumed by different add-ons, including [Grafana](https://grafana.com/) for analytics and monitoring, and [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) for handling alerts.

## Tracing

The tracing in Kyma uses the [Jaeger](https://github.com/jaegertracing) distributed tracing system. Use it to analyze performance by scrutinizing the path of the requests sent to and from your service. This information helps you optimize the latency and performance of your solution.

## API Gateway

The API Gateway aims to provide a set of functionalities which allow developers to expose, secure, and manage their APIs in an easy way. The main element of the API Gateway is the API Gateway Controller which exposes services in Kyma.

## Logging

Logging in Kyma uses [Loki](https://github.com/grafana/loki), a Prometheus-like log management system.

## Console

The Console is a web-based administrative UI for Kyma. It uses the [Luigi framework](https://github.com/SAP/luigi) to allow you to seamlessly extend the UI content with custom micro frontends. The Console has a dedicated Console Backend Service which provides a tailor-made API for each view of the Console UI.

## Rafter

Rafter is a solution for storing and managing different types of assets, such as documents, files, images, API specifications, and client-side applications. It uses [MinIO](https://min.io/) as object storage and relies on Kubernetes custom resources (CRs). The custom resources are managed by a controller that communicates through MinIO Gateway with external cloud providers.

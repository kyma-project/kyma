---
title: Components
type: Details
---

Kyma is built on the foundation of the best and most advanced open-source projects which make up the components readily available for customers to use.
This section describes the Kyma components.

## Security

Kyma security enforces role-based access control (RBAC) in the cluster. [Dex](https://github.com/dexidp/dex) handles the identity management and identity provider integration. It allows you to integrate any [OpenID Connect](https://openid.net/connect/) or SAML2-compliant identity provider with Kyma using [connectors](https://github.com/dexidp/dex#connectors). Additionally, Dex provides a static user store which gives you more flexibility when managing access to your cluster.

## Service Catalog

The Service Catalog lists all of the services available to Kyma users through the registered [Service Brokers](/components/service-catalog/#service-brokers-service-brokers). Use the Service Catalog to provision new services in the
Kyma [Kubernetes](https://kubernetes.io/) cluster and create bindings between the provisioned service and an application.

## Helm Broker

The Helm Broker is a Service Broker which runs in the Kyma cluster and deploys Kubernetes native resources using [Helm](https://github.com/kubernetes/helm) and Kyma bundles. A bundle is an abstraction layer over a Helm chart which allows you to represent it as a ClusterServiceClass in the Service Catalog. Use bundles to install the [GCP Broker](/components/service-catalog#service-brokers-gcp-broker), [Azure Service Broker](/components/service-catalog#service-brokers-azure-service-broker) and the [AWS Service Broker](/components/service-catalog#service-brokers-aws-service-broker) in Kyma.

## Application Connector

The Application Connector is a proprietary Kyma solution. This endpoint is the Kyma side of the connection between Kyma and the external solutions. The Application Connector allows you to register the APIs and the Event Catalog, which lists all of the available events, of the connected solution. Additionally, the Application Connector proxies the calls from Kyma to external APIs in a secure way.

## Event Bus

Kyma Event Bus receives Events from external solutions and triggers the business logic created with lambda functions and services in Kyma. The Event Bus is based on the [NATS Streaming](https://nats.io/) open source messaging system for cloud-native applications.

## Service Mesh

The Service Mesh is an infrastructure layer that handles service-to-service communication, proxying, service discovery, traceability, and security, independently of the code of the services. Kyma uses the [Istio](https://istio.io/) Service Mesh that is customized for the specific needs of the implementation.

## Serverless

The Kyma Serverless component allows you to reduce the implementation and operation effort of an application to the absolute minimum. Kyma Serverless provides a platform to run lightweight functions in a cost-efficient and scalable way using JavaScript and Node.js. Kyma Serverless is built on the [Kubeless](http://kubeless.io/) framework, which allows you to deploy lambda functions, and uses the [NATS](https://nats.io/) messaging system that monitors business events and triggers functions accordingly.

## Monitoring

Kyma comes bundled with tools that give you the most accurate and up-to-date monitoring data. [Prometheus](https://prometheus.io/) open source monitoring and alerting toolkit provides this data, which is consumed by different add-ons, including [Grafana](https://grafana.com/) for analytics and monitoring, and [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) for handling alerts.

## Tracing

The tracing in Kyma uses the [Jaeger](https://github.com/jaegertracing) distributed tracing system. Use it to analyze performance by scrutinizing the path of the requests sent to and from your service. This information helps you optimize the latency and performance of your solution.

## API Gateway

The API Gateway aims to provide a set of functionalities which allow developers to expose, secure, and manage their APIs in an easy way. The main element of the API Gateway is the API Gateway Controller which exposes services in Kyma.

>**CAUTION:** The API Gateway implementation that uses the API Controller and the Api custom resource is **deprecated** until all of its functionality is covered by the `v2` implementation. The services you exposed and secured so far do not require any action as two implementations co-exist in the cluster. When you expose new services and functions secured with OAuth2, use the `v2` implementation. For more information, read [this](/components/api-gateway-v2#overview-overview) documentation.

## Logging

Logging in Kyma uses [Loki](https://github.com/grafana/loki), a Prometheus-like log management system.

## Backup

Kyma integrates with [Velero](https://github.com/heptio/velero/) to provide backup and restore capabilities for Kubernetes cluster resources. Once backed up, Velero stores the resources in buckets of [supported cloud providers](https://velero.io/docs/v1.0.0/support-matrix/).

## Console

The Console is a web-based administrative UI for Kyma. It uses the [Luigi framework](https://github.com/SAP/luigi) to allow you to seamlessly extend the UI content with custom micro frontends. The Console has a dedicated Console Backend Service which provides a tailor-made API for each view of the Console UI.

## Asset Store

The Asset Store is a flexible, scalable, multi-cloud, and location-independent Kubernetes-native solution for storing assets, such as documents, files, images, API specifications, and client-side applications. The Asset Store consists of [MinIO](https://min.io/) and custom resources. The custom resources are managed by a controller that communicates through MinIO Gateway with external cloud providers.

## Headless CMS

The Headless CMS is a Kubernetes-native Content Management System (CMS) that provides a way of storing and managing raw content, and exposing it through an API. The Headless CMS is an abstraction on top of the Asset Store which provides a data layer. The Headless CMS has no dedicated presentation layer but it is integrated with the Console UI that consumes content stored using Headless CMS.

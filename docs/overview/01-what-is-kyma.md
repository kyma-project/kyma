---
title: What is Kyma?
type: Overview
---

<!--

- In a nutshell - rewrite -> Overview (PB)
- Main features - remove (reuse some parts in the Overview)

-->

Kyma allows you to extend applications with microservices and Functions. First, connect your application to a Kubernetes cluster and expose the application's API or events securely. Then, implement the business logic you require by creating microservices or Functions and triggering them to react to particular events or calls to your application's API. To limit the time spent on coding, use the built-in cloud services from Service Catalog, exposed by open service brokers from such cloud providers as GCP, Azure, and AWS.

Kyma comes equipped with these out-of-the-box functionalities:

- Service-to-service communication and proxying (Istio-based [Service Mesh](/components/service-mesh/#overview-overview))
- Built-in [monitoring](/components/monitoring/#overview-overview), [tracing](/components/tracing/#overview-overview), and [logging](/components/logging/#overview-overview) (Grafana, Prometheus, Jaeger, Loki, Kiali)
- Secure [authentication and authorization](/components/security/#overview-overview) (Dex, Ory, Service Identity, TLS, Role Based Access Control)
- The catalog of services to choose from ([Service Catalog](/components/service-catalog/#overview-service-catalog), [Service Brokers](/components/service-catalog/#overview-service-brokers))
- The development platform to run lightweight Functions in a cost-efficient and scalable way ([Serverless](/components/serverless/#overview-overview))
- The endpoint to register Events and APIs of external applications ([Application Connector](/components/application-connector/#overview-overview))
- Secure API exposure ([API Gateway](/components/api-gateway/#overview-overview))
- The messaging channel to receive Events, enrich them, and trigger business flows using Functions or services ([Eventing](/components/eventing/#overview-overview), NATS)
- CLI supported by the intuitive UI ([Console](/components/console/#overview-overview))
- Asset management and storing tool ([Rafter](/components/rafter/#overview-overview), MinIO)
- Backup of Kyma clusters ([Kyma Backup](/root/kyma/#installation-back-up-kyma))


Major open-source and cloud-native projects, such as Istio, NATS, Serverless, and Prometheus, constitute the cornerstone of Kyma. Its uniqueness, however, lies in the "glue" that holds these components together. Kyma collects those cutting-edge solutions in one place and combines them with the in-house developed features that allow you to connect and extend your enterprise applications easily and intuitively.

Kyma allows you to extend and customize the functionality of your products in a quick and modern way, using serverless computing or microservice architecture. The extensions and customizations you create are decoupled from the core applications, which means that:
- Deployments are quick.
- Scaling is independent from the core applications.
- The changes you make can be easily reverted without causing downtime of the production system.

Last but not least, Kyma is highly cost-efficient. All Kyma native components and the connected open-source tools are written in Go. It ensures low memory consumption and reduced maintenance costs compared to applications written in other programming languages such as Java.

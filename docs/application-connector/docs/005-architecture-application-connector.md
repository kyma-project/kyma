---
title: Application Connector components
type: Architecture
---

The Application Connector consists of the following components:

* **Nginx Ingress Controller**
* **Connector Service** 
* **Metadata Service**
* **Event Service**
* **Proxy Service**
* **Remote Environments**
* **Minio bucket**
* **Kubernetes Secrets**
* **Kubernetes Services**


![Architecture Diagram](assets/001-application-connector.png)


### Nginx Ingress Controller

The Nginx Ingress Controller controller exposes the Application Connector to the outside world. 
It handles the routing to the other components and secures communication using the client certificates.

The detailed description can be found here: [Nginx Ingress Controller](./006-architecture-ingress-gateway.md)


### Connector Service

The Connector Service is handling the exchange of the client certificates for a given Remote Environment and also returns the Metadata service and Event service endpoints.
The Connector Service is signing each client certificate with use of server side certificate which is stored in Kubernetes secrets.

The detailed description can be found here: [Connector Service](TBD)


### Metadata Service

The Metadata Service stores all registered APIs and Event Catalog which are exposed by the connected system.
The APIs and Event catalogs metadata are stored in Remote Environment Custom Resource and the documentation is stored in Minio bucket.

The Kubernetes services are created for each registered API and a new Service classes in Service Catalog is registered.
Therefore, the services and Lambda function within Kyma can access the external API using the Service Catalog binding and the Proxy service.
 
The API can be registered together with an OAuth credentials which are stored in Kubernetes secrets.

The detailed description can be found here: [Metadata Service](TBD)


### Proxy service

The Proxy service is sending events to the Kyma Eventbus. The main purpose of the service to enrich the events with additional metadata indicating the source of the event.
That allows routing the events to Lambda functions and Services based on Remote Environment from which they are coming.

The detailed description can be found here: [Proxy Service](TBD)


### Remote Environments

Remote Environment represents the external system which is connected to Kyma.


### Minio bucket

Minio bucket is the storag for documentation of the registered APIs and Event catalogs.

### Kubernetes Secret

Kubernetes Secret is a Kubernetes Object which stores sensitive data, like OAuth credentials

### Kubernetes Services

Kubernetes services are used for managing an access to the external API over the Proxy service from the Lambda functions and services deployed in Kyma.



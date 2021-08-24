---
title: Glossary
---

Here's a list of the most commonly used terms that you may come across when browsing through the Kyma documentation:

>**TIP:** The terms in the list are ordered alphabetically.

| Term |  Description | Useful links |
| ---- | ------------ | ------------ |
| Credentials/Secrets | Sensitive data to call the service, connect to it, and authenticate it.  |       |
| Function            | A simple code snippet that you can run without provisioning or managing servers. It implements the exact business logic you define. A Function is based on the Function custom resource (CR) and can be written in either Node.js or Python. A Function can perform a business logic of its own. You can also bind it to an instance of a service and configure it to be triggered whenever it receives a particular event type from the service or a call is made to the service's API. Functions are executed only if they are triggered by an event or an API call. | [What is Serverless in Kyma?](./01-overview/main-areas/serverless)      |
| Kyma cluster | A [Kubernetes cluster](https://kubernetes.io/docs/reference/glossary/?fundamental=true#term-cluster) with Kyma installed.  | [Kubernetes components](https://kubernetes.io/docs/concepts/overview/components/) |
| Kyma Dashboard | Kyma Dashboard is a web-based UI for managing resources within Kyma or any Kubernetes cluster. | [What are the UIs available in Kyma?](./01-overview/main-areas/ui)
| Microservice | An architectural variant for extensions or applications, where you separate the tasks into smaller pieces that interact with each other as loosely coupled, independently deployable units of code. A failing microservice should not cause your whole application to fail. Microservices are packed in a container that is always running; it's idling if there is no load. The microservice should always be reachable even when the Pods move around. Microservices typically communicate through APIs. |       |
| Namespace | Namespaces organize objects in a cluster and provide a way to divide cluster resources. This way, several users can share a cluster but have access to resources restricted to specified Namespaces only. This increases the security and organization of your cluster by dividing it into smaller units. Access to Namespaces in the Kyma environment depends on your permissions. | [Kubernetes Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/) |
| Role | Access to every cluster is managed by the roles assigned. Roles give the assigned users a different level of permissions suitable for different purposes. | [Authorization in Kyma](./04-operation-guides/security/sec-03-authorization-in-kyma.md)      |
| Service operators | Service Management in Kyma uses service operators. A service operator is a piece of software that provides a set of all necessary resources (such as CustomResourceDefinitions and controllers) needed to provision third-party services in your Kubernetes cluster. | [What is Service Management in Kyma?](./01-overview/main-areas/service-management)

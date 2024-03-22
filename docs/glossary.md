# Glossary

Here's a list of the most commonly used terms that you may come across when browsing through the Kyma documentation:

> [!TIP]
> The terms in the list are ordered alphabetically.

| Term |  Description | Useful links |
| ---- | ------------ | ------------ |
| Application | An external solution connected to Kyma through Application Connector.   <br><br> <div style="background-color:#ffddd3; padding:9px;"> **Warning** <br> Don't confuse it with `application`, which is the term used for a microservice deployed on Kyma or in a general sense for software.</div>  | [Application custom resource](https://kyma-project.io/#/application-connector-manager/user/resources/06-10-application)      |
| Credentials/Secrets | Sensitive data to call the service, connect to it, and authenticate it.  |       |
| Custom resource | A custom resource (CR) allows you to extend the Kubernetes API to cover use cases that are not directly covered by core Kubernetes.  | [Kubernetes - Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) & [Custom resources provided by Kyma](./05-technical-reference/00-custom-resources)  |
| CustomResourceDefinition (CRD) | An object used to define a custom resource. | [CRDs of custom resources provided by Kyma](https://github.com/kyma-project/kyma/tree/main/installation/resources/crds)      |
| Deployment | Deployment is a Kubernetes object that represents a replicated application running in your cluster.       | [Kubernetes - Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)        |
| Function | A simple code snippet that you can run without provisioning or managing servers. It implements the exact business logic you define. A Function is based on the Function custom resource (CR) and can be written in either Node.js or Python. A Function can perform a business logic of its own. You can also bind it to an instance of a service and configure it to be triggered whenever it receives a particular event type from the service or a call is made to the service's API. Functions are executed only if they are triggered by an event or an API call. | [What is Serverless in Kyma?](https://kyma-project.io/#/serverless-manager/user/README)      |
| Kyma cluster | A Kubernetes cluster with Kyma installed.  | [Kubernetes components](https://kubernetes.io/docs/concepts/overview/components/) |
| Kyma dashboard | Kyma dashboard is a web-based UI for managing resources within Kyma or any Kubernetes cluster. | [What are the UIs available in Kyma?](./01-overview/ui) |
| Microservice | An architectural variant for extensions or applications, where you separate the tasks into smaller pieces that interact with each other as loosely coupled, independently deployable units of code. A failing microservice should not cause your whole application to fail. Microservices are packed in a container that is always running; it's idling if there is no load. The microservice should always be reachable even when the Pods move around. Microservices typically communicate through APIs. |       |
| Namespace | Namespaces organize objects in a cluster and provide a way to divide cluster resources. This way, several users can share a cluster but have access to resources restricted to specified namespaces only. This increases the security and organization of your cluster by dividing it into smaller units. Access to namespaces in the Kyma environment depends on your permissions. | [Kubernetes - Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/) |
| Pod | Pod is the smallest and simplest Kubernetes object that represents a set of containers running in your cluster.  | |
| Role | Access to every cluster is managed by the roles assigned. Roles give the assigned users a different level of permissions suitable for different purposes. | [Authorization in Kyma](./04-operation-guides/security/sec-02-authorization-in-kyma.md)      |
| Service | A Service in Kubernetes is an abstract way to expose an application running on a set of Pods as a network service. <br><br> In Kyma documentation, we use `Service` to refer to the Kubernetes term, and `service` to refer to a software functionality in general. | [Kubernetes -  Service](https://kubernetes.io/docs/concepts/services-networking/service/)    |
| Service operator | Service Management in Kyma uses service operators. A service operator is a piece of software that provides a set of all necessary resources (such as CustomResourceDefinition and controllers) needed to provision third-party services in your Kubernetes cluster. | [OperatorHub.io](https://operatorhub.io/) & [SAP BTP Operator Module](https://kyma-project.io/#/btp-manager/user/README) |

> [!TIP]
> To learn the basic Kubernetes terminology, read the [Kubernetes glossary](https://kubernetes.io/docs/reference/glossary).

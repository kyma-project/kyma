---
title: Architecture
---

This diagram illustrates the workflow that leads to exposing a service in Kyma:

![service-exposure-flow](./assets/001-api-gateway-flow.svg)

- **API Gateway Controller** is a component responsible for exposing services. The API Gateway Controller is an application deployed in the `kyma-system` Namespace, implemented according to the [Kubernetes Operator](https://coreos.com/blog/introducing-operators.html) principles. The API Gateway Controller listens for newly created custom resources (CR) that follow the set `apirule.gateway.kyma-project.io` CustomResourceDefinition (CRD), which describes the details of exposing services in Kyma.

- **Istio Virtual Service** is used to specify the services that are visible outside the cluster. The API Gateway Controller creates a Virtual Service for the hostname defined in the `apirule.gateway.kyma-project.io` CRD. The convention is to create a hostname using the name of the service as the subdomain, and the domain of the Kyma cluster. To learn more about the Istio Virtual Service concept, read this [Istio documentation](https://kubernetes.io/docs/concepts/services-networking/ingress/).
To get the list of Virtual Services in Kyma, run:

  ```
  kubectl get virtualservices.networking.istio.io --all-namespaces
  ```

- **Oathkeeper Access Rule** allows operators to specify authentication requirements for a service. It is an optional resource, created only when the CR specifies the desired authentication method, the trusted token issuer, allowed methods and paths, and required scopes. To learn more about Oathkeeper Access Rules, read [this](https://www.ory.sh/docs/oryos.10/oathkeeper/api-access-rules) document.

To get the list of Istio Authentication Policies created in Kyma, run:

  ```
  kubectl get rules.oathkeeper.ory.sh --all-namespaces
  ```

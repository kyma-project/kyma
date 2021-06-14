---
title: Serverless
---

"Serverless" refers to an architecture in which the infrastructure of your applications is managed by cloud providers. Contrary to its name, a serverless application does require a server but it doesn't require you to run and manage it on your own. Instead, you subscribe to a given cloud provider, such as AWS, Azure, or GCP, and pay a subscription fee only for the resources you actually use. Since the resource allocation can be dynamic and depends on your current needs, the serverless model is particularly cost-effective when you want to implement a certain logic that is triggered on demand. Simply, you get your things done and don't pay for the infrastructure that stays idle.

Kyma offers a service (known as "functions-as-a-service" or "FaaS") that provides a platform on which you can build, run, and manage serverless applications in Kubernetes. These applications are called **Functions** and they are based on [Function custom resource (CR)](../../../05-technical-references/06-custom-resources/svls-01-function.md) objects. They contain simple code snippets that implement a specific business logic. For example, you can define that you want to use a Function as a proxy that saves all incoming event details to an external database.

Such a Function can be:

- Triggered by other workloads in the cluster (in-cluster events) or business events coming from external sources. You can subscribe to them using a [Subscription CR](../../../05-technical-references/06-custom-resources/evnt-01-subscription.md).
- Exposed to an external endpoint (HTTPS). With an [APIRule CR](../../../05-technical-references/06-custom-resources/apig-01-apirule.md), you can define who can reach the endpoint and what operations they can perform on it.

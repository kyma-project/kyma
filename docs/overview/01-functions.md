---
title: Functions
type: Overview
---

"Serverless" refers to an architecture in which the infrastructure of your applications is managed by cloud providers. Contrary to its name, a serverless application does require a server but it doesn't require you to run and manage it on your own. Instead, you subscribe to a given cloud provider, such as AWS, Azure or GCP, and pay a subscription fee only for the resources you actually use. Since the resource allocation can be dynamic and depends on your current needs, the serverless model is particularly cost-effective when you want to implement a certain logic that is triggered on demand. Simply, you get your things done and don't pay for the infrastructure that sits idle.

Kyma offers a service (known as "functions-as-a-service" or "FaaS") that provides a platform on which you can build, run, and manage serverless applications in Kubernetes. These applications are called **Functions** that are based on [Function custom resource (CR)](#custom-resource-function) objects. They contain simple code snippets that implement a specific business logic. For example, you can define that you want to use the Function as a proxy that saves all incoming event details to an external database.

Such a Function can be:

- Triggered by other Functions or events coming from external sources. You can subscribe to them using Subscription CRs.
- Exposed to an external endpoint (HTTPS). With APIRule CRs, you can also define who can reach the endpoint and what operations they can perform on it.

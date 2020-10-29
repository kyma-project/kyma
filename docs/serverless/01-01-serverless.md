---
title: Overview
---

"Serverless" refers to an architecture in which the infrastructure of your applications is managed by cloud providers. Contrary to its name, a serverless application does require a server but it doesn't require you to run and manage it on your own. Instead, you subscribe to a given cloud provider, such as AWS, Azure or GCP, and pay a subscription fee only for the resources you actually use. Since the resource allocation can be dynamic and depends on your current needs, the serverless model is particularly cost-effective when you want to implement a certain logic that is triggered on demand. Simply, you get your things done and don't pay for the infrastructure that sits idle.

Similarly to cloud providers, Kyma offers a service (known as "functions-as-a-service" or "FaaS") that provides a platform on which you can build, run, and manage serverless applications in Kubernetes. These applications are called **Functions** and are based on the Function CR objects. They are simple code snippets that implement the exact business logic you define in them. After you create a Function, you can:

- Configure it to be triggered by events coming from external sources to which you subscribe.
- Expose it to an external endpoint (HTTPS).

> **CAUTION:** In its default configuration, Serverless uses persistent volumes as the internal registry to store Docker images for Functions. The default storage size of a single volume is 20 GB. This internal registry is suitable for local development. For production purposes, we recommend to use an [external Docker registry](#tutorials-set-an-external-docker-registry).

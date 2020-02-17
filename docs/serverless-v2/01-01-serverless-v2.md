---
title: Overview
---

>**CAUTION:** This implementation will soon replace Serverless that is based on [Kubeless](https://github.com/kubeless/kubeless). Serverless v2 relies on [Knative Serving](https://knative.dev/docs/serving/) for deploying and managing functions, and [Tekton Pipelines](https://github.com/tektoncd/pipeline) for creating and running cluster-wide processes. Consider this implementation as experimental as it is still under development.

"Serverless" refers to an architecture in which the infrastructure of your applications is managed by cloud providers. Contrary to its name, serverless applications do require a server but they don't require you to run and manage it on your own. Instead, you subscribe to a given cloud provider, such as AWS, Azure or GCP, and pay a subscription fee only for the resources you actually use. Since the resource allocation can be dynamic and depends on your current needs, the serverless model is particularly cost-effective when you want to implement a certain logic that is triggered on demand. Simply, you get your things done and don't pay for the infrastructure that sits idle.

Similarly to cloud providers, Kyma offers a service (known as "functions-as-a-service" or "FaaS") that provides a platform on which you can build, run, and manage serverless applications in Kubernetes. These applications are called **lambda functions** (based on the Function CR object) or **lambdas** and are simple code snippets that implement the exact business logic you define in them. After you create a lambda, you can:

- Configure it to be triggered by events coming from external sources to which you subscribe.
- Expose it to an external endpoint (HTTPS).

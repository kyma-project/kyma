---
title: Overview
---

>**CAUTION:** This implementation will soon replace Serverless that is based on [Kubeless](https://github.com/kubeless/kubeless). Serverless v2 replies on [Knative Serving](https://knative.dev/docs/serving/) for deploying and managing functions, and [Tekton](https://github.com/tektoncd/pipeline) as a pipeline for creating and running on-cluster processes. Consider this implementation as experimental as it is still under development.

"Serverless" refers to an architecture in which the infrastructure of your applications is managed by cloud providers. Contrary to its name, serverless applications do require a server but they don't require you to run and manage it on your own. Instead, you subscribe to a given cloud provider, such as AWS, Azure or GCP, and pay a subscription fee only for the resources you actually use. Since the resource allocation can be dynamic and depend on your current needs, the serverless model is particularly cost-effective when you want to perform a certain logic that is triggered on-demand. Simply, you get your things done and don't pay for the infrastructure that sits idle.

Kyma provides a service (known as "functions-as-a-service" or FaaS") that offers a platform on which you can build, run, and manage serverless applications called **functions** (based on a Function CR object) or **lambdas**. These JavaScript code snippets perform the business logic you define in them. After you create a lambda, you can:

- Configure it to be triggered by incoming events from external sources to which you subscribe.
- Expose it to an external endpoint (HTTPS).

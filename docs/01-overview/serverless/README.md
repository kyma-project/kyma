---
title: What is Serverless in Kyma?
---

Serverless in Kyma is an area that:

- Ensures quick deployments following a Function approach
- Enables scaling independent of the core applications
- Gives a possibility to revert changes without causing production system downtime
- Supports the complete asynchronous programming model
- Offers loose coupling of Event providers and consumers
- Enables flexible application scalability and availability

Serverless in Kyma allows you to reduce the implementation and operation effort of an application to the absolute minimum. It provides a platform to run lightweight Functions in a cost-efficient and scalable way using JavaScript and Node.js. Serverless in Kyma relies on Kubernetes resources like [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/), [Services](https://kubernetes.io/docs/concepts/services-networking/service/) and [HorizontalPodAutoscalers](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) for deploying and managing Functions and [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) for creating Docker images.

"Serverless" refers to an architecture in which the infrastructure of your applications is managed by cloud providers. Contrary to its name, a serverless application does require a server but it doesn't require you to run and manage it on your own. Instead, you subscribe to a given cloud provider, such as AWS, Azure, or GCP, and pay a subscription fee only for the resources you actually use. Because the resource allocation can be dynamic and depends on your current needs, the serverless model is particularly cost-effective when you want to implement a certain logic that is triggered on demand. Simply, you get your things done and don't pay for the infrastructure that stays idle.

Kyma offers a service (known as "functions-as-a-service" or "FaaS") that provides a platform on which you can build, run, and manage serverless applications in Kubernetes. These applications are called **Functions** and they are based on [Function custom resource (CR)](../../05-technical-reference/00-custom-resources/svls-01-function.md) objects. They contain simple code snippets that implement a specific business logic. For example, you can define that you want to use a Function as a proxy that saves all incoming event details to an external database.

Such a Function can be:

- Triggered by other workloads in the cluster (in-cluster events) or business events coming from external sources. You can subscribe to them using a [Subscription CR](../../05-technical-reference/00-custom-resources/evnt-01-subscription.md).
- Exposed to an external endpoint (HTTPS). With an [APIRule CR](../../05-technical-reference/00-custom-resources/apix-01-apirule.md), you can define who can reach the endpoint and what operations they can perform on it.

# From code to Fucation

Pick the programming language for the Function and decide where you want to keep the source code. Serverless will create the workload out of it for you.

## Runtimes

Functions support multiple languages by using the underlying execution environments known as runtimes. Currently, you can create both Node.js and Python Functions in Kyma.

>**TIP:** See [sample Functions](../../05-technical-reference/svls-01-sample-functions.md) for each available runtime.

## Source code

You can also choose where you want to keep your Function's source code and dependencies. You can either place them directly in the Function CR under the **spec.source** and **spec.deps** fields as an **inline Function**, or store the code and dependencies in a public or private Git repository (**Git Functions**). Choosing the second option ensures your Function is versioned and gives you more development freedom in the choice of a project structure or an IDE.

>**TIP:** Read more about [Git Functions](../../05-technical-reference/svls-04-git-source-type.md).

# Container registries

By default, Serverless uses PersistentVolume (PV) as the internal registry to store Docker images for Functions. The default storage size of a single volume is 20 GB. This internal registry is suitable for local development.

If you use Serverless for production purposes, it is recommended that you use an external registry, such as Docker Hub, Google Container Registry (GCR), or Azure Container Registry (ACR).

Serverless supports two ways of connecting to an external registry:

- [You can set up an external registry before installation](../../03-tutorials/00-serverless/svls-07-set-external-registry.md).

  In this scenario, you can use Kyma overrides to change the default values supplied by the installation mechanism.

- [You can switch to an external registry at runtime](../../03-tutorials/00-serverless/svls-08-switch-to-external-registry.md).

  In this scenario, you can change the registry on the fly, with Kyma already installed on your cluster. This option gives you way more flexibility and control over the choice of an external registry for your Functions.

>**TIP:** For details, read about [switching registries at runtime](../../05-technical-reference/svls-03-switching-registries.md).

# Development toolkit

To start developing your first Functions, you need:

- Self-hosted **Kubernetes cluster** and the **KUBECONFIG** file to authenticate to the cluster
- **Kyma** as the platform for managing the Function-related workloads
- [**Docker**](https://www.docker.com/) as the container runtime
- [**kubectl**](https://kubernetes.io/docs/reference/kubectl/kubectl/), the Kubernetes command-line tool, for running commands against clusters
- Development environment of your choice:
   - **Kyma CLI** to easily initiate inline Functions or Git Functions locally, run, test, and later apply them on the clusters
   - **Node.js** (v14 or v16) or **Python** (v3.9)
   - **IDE** as the source code editor
   - **Kyma Dashboard** to manage Functions and related workloads through the graphical user interface

# Security considerations

To eliminate potential security risks when using Functions, bear in mind these few facts:

- Kyma provides base images for serverless runtimes. Those default runtimes are maintained with regards to commonly known security advisories. It is possible to use a custom runtime image (see this [tutorial](../../03-tutorials/00-serverless/svls-13-override-runtime-image.md)). In such a case, you are responsible for security compliance and assessment of exploitability of any potential vulnerabilities of the custom runtime image.

- Kyma does not run any security scans against Functions and their images. Before you store any sensitive data in Functions, consider the potential risk of data leakage.

- Kyma does not define any authorization policies that would restrict Functions' access to other resources within the Namespace. If you deploy a Function in a given Namespace, it can freely access all events and APIs of services within this Namespace.

- Since Kubernetes is [moving from PodSecurityPolicies to PodSecurity Admission Controller](https://kubernetes.io/docs/tasks/configure-pod-container/migrate-from-psp/), Kyma Functions require running in Namespaces with the `baseline` Pod security level. The `restricted` level is not currently supported due to the requirements of the Function building process.

- Kyma Serverless components can run with the PodSecurity Admission Controller support in the `restricted` Pod security level when using an external registry. When the Internal Docker Registry is enabled, the Internal Registry DaemonSet requires elevated privileges to function correctly, exceeding the limitations of both the `restricted` and `baseline` levels.

- All administrators and regular users who have access to a specific Namespace in a cluster can also access:

  - Source code of all Functions within this Namespace
  - Internal Docker registry that contains Function images
  - Secrets allowing the build Job to pull and push images from and to the Docker registry (in non-system Namespaces)

# Limitations

## Controller limitations

Serverless controller does not serve time-critical requests from users.
It reconciles Function custom resources (CR), stored at the Kubernetes API Server, and has no persistent state on its own.

Serverless controller doesn't build or serve Functions using its allocated runtime resources. It delegates this work to the dedicated Kubernetes workloads. It schedules (build-time) jobs to build the Function Docker image and (runtime) Pods to serve them once they are built. 
Refer to the [architecture](../../05-technical-reference/00-architecture/svls-01-architecture.md) diagram for more details.

Having this in mind Serverless controller does not require horizontal scaling.
It scales vertically up to the `160Mi` of memory and `500m` of CPU time.

## Limitation for the number of Functions

There is no upper limit of Functions that can be run on Kyma (similar to Kubernetes workloads in general). Once a user defines a Function, its build jobs and runtime Pods will always be requested by Serverless controller. It's up to Kubernetes to schedule them based on the available memory and CPU time on the Kubernetes worker nodes. This is determined mainly by the number of the Kubernetes worker nodes (and the node auto-scaling capabilities) and their computational capacity.

## Build phase limitation:

The time necessary to build Function depends on:
 - selected [build profile](../../05-technical-reference/svls-08-available-presets.md#build-jobs-resources) that determines the requested resources (and their limits) for the build phase 
 - number and size of dependencies that must be downloaded and bundled into the Function image.
 - cluster nodes specification (see the note with reference specification at the end of the article)

<div tabs name="build" group="function-build-times">
  <details>
  <summary label="nodejs">
  Node.js 
  </summary>

|                 | local-dev | no profile (no limits for resource) |
|-----------------|-----------|-------------------------------------|
| no dependencies | 24 sec    | 15 sec                              |
| 2 dependencies  | 26 sec    | 16 sec                              |

  </details>
  <details>
  <summary label="python">
  Python
  </summary>

|                 | local-dev | no profile (no limits for resource) |
|-----------------|-----------|-------------------------------------|
| no dependencies | 30 sec    | 16 sec                              |
| 2 dependencies  | 32 sec    | 20 sec                              |

  </details>
</div>

The shortest build time (the limit) is approximately 15 seconds and requires no limitation of the build job resources and a minimum number of dependencies that are pulled in during the build phase.

Running multiple Function build jobs at once (especially with no limits) may drain the cluster resources. To mitigate such risk, there is an additional limit of 5 simultaneous Function builds. If a sixth one is scheduled, it is built once there is a vacancy in the build queue.

This limitation is configurable using [`containers.manager.envs.functionBuildMaxSimultaneousJobs`](../../05-technical-reference/00-configuration-parameters/svls-01-serverless-chart.md#configurable-parameters).

## Runtime phase limitations
In the runtime, the Functions serve user-provided logic wrapped in the WEB framework (`express` for Node.js and `bottle` for Python). Taking the user logic aside, those frameworks have limitations and depend on the selected [runtime profile](../../05-technical-reference/svls-08-available-presets.md#functions-resources) and the Kubernetes nodes specification (see the note with reference specification at the end of this article).

The following describes the response times of the selected runtime profiles for a "hello world" Function requested at 50 requests/second. This describes the overhead of the serving framework itself. Any user logic added on top of that will add extra milliseconds and must be profiled separately.


<div tabs name="steps" group="function-response-times">
  <details>
  <summary label="nodejs">
  Node.js 
  </summary>
|                               | XL     | L      | M      | S      | XS      |
|-------------------------------|--------|--------|--------|--------|---------|
| response time [avarage]       | ~13ms  | 13ms   | ~15ms  | ~60ms  | ~400ms  |
| response time [95 percentile] | ~20ms  | ~30ms  | ~70ms  | ~200ms | ~800ms  |
| response time [99 percentile] | ~200ms | ~200ms | ~220ms | ~500ms | ~1.25ms |


  </details>
  <details>
  <summary label="python">
  Python
  </summary>

|                               | XL     | L      | M      | S      |
|-------------------------------|--------|--------|--------|--------|
| response time [avarage]       | ~11ms  | 12ms   | ~12ms  | ~14ms  |
| response time [95 percentile] | ~25ms  | ~25ms  | ~25ms  | ~25ms  |
| response time [99 percentile] | ~175ms | ~180ms | ~210ms | ~280ms |
  </details>
</div>

Obviously, the bigger the runtime profile, the more resources are available to serve the response quicker. Consider these limits of the serving layer as a baseline - as this does not take your Function logic into account.

### Scaling

Function runtime Pods can be scaled horizontally from zero up to the limits of the available resources at the Kubernetes worker nodes.
See the [Use external scalers](../../03-tutorials/00-serverless/svls-15-use-external-scalers.md) tutorial for more information.

## In-cluster Docker registry limitations

Serverless comes with an in-cluster Docker registry for the Function images.
This registry is only suitable for development because of its [limitations](../serverless/svls-03-container-registries.md), i.e.:
 - Registry capacity is limited to 20GB
 - There is no image lifecycle management. Once an image is stored in the registry, it stays there until it is manually removed.

> **NOTE:** All measurements were done on Kubernetes with five AWS worker nodes of type `m5.xlarge` (four CPU 3.1 GHz x86_64 cores, 16 GiB memory).

# Useful links

If you're interested in learning more about the Serverless area, follow these links to:

- Perform some simple and more advances tasks:

  - Create an [inline](../../03-tutorials/00-serverless/svls-01-create-inline-function.md) or a [Git](../../03-tutorials/00-serverless/svls-02-create-git-function.md) Function
  - [Expose the Function](../../03-tutorials/00-serverless/svls-03-expose-function.md)
  - [Manage Functions through Kyma CLI](../../03-tutorials/00-serverless/svls-04-manage-functions-with-kyma-cli.md)
  - [Debug a Function](../../03-tutorials/00-serverless/svls-05-debug-function.md)
  - [Synchronize Functions in a GitOps fashion](../../03-tutorials/00-serverless/svls-06-sync-function-with-gitops.md)
  - [Set an external Docker registry](../../03-tutorials/00-serverless/svls-07-set-external-registry.md) for your Function images and [switch between registries at runtime](../../03-tutorials/00-serverless/svls-08-switch-to-external-registry.md)
  - [Log into a private package registry](../../03-tutorials/00-serverless/svls-09-log-into-private-packages-registry.md)

- Troubleshoot Serverless-related issues when:

   - [Functions won't build](../../04-operation-guides/troubleshooting/serverless/svls-01-cannot-build-functions.md)
   - [Container fails](../../04-operation-guides/troubleshooting/serverless/svls-02-failing-function-container.md)
   - [Debugger stops](../../04-operation-guides/troubleshooting/serverless/svls-03-function-debugger-in-strange-location.md)

- Analyze Function specification and configuration files:

  - [Function](../../05-technical-reference/00-custom-resources/svls-01-function.md) custom resource
  - [`config.yaml` file](../../05-technical-reference/svls-06-function-configuration-file.md) in Kyma CLI
  - [Function specification details](../../05-technical-reference/svls-07-function-specification.md)

- Understand technicalities behind Serverless implementation:

  - [Serverless architecture](../../05-technical-reference/00-architecture/svls-01-architecture.md) and [Function processing](../../05-technical-reference/svls-02-function-processing-stages.md)
  - [Switching registries](../../05-technical-reference/svls-03-switching-registries.md)
  - [Git source type](../../05-technical-reference/svls-04-git-source-type.md)
  - [Exposing Functions](../../05-technical-reference/svls-05-exposing-functions.md)
  - [Available presets](../../05-technical-reference/svls-08-available-presets.md)
  - [Environment variables in Functions](../../05-technical-reference/00-configuration-parameters/svls-02-environment-variables.md)

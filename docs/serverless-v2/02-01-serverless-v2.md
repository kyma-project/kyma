---
title: Architecture
---

Serverless v2 relies on [Knative Serving](https://knative.dev/docs/serving/) for deploying and managing functions and [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) for creating Docker images. See how these and other resources process a lambda within a Kyma cluster:

![Serverless architecture](./assets/serverless-architecture.svg)

1. Create a lambda either through the UI or by applying a Function custom resource (CR). This CR contains the lambda definition (business logic that you want to execute) and information on the environment on which it should run.

    >**NOTE:** Function Controller currently supports Node.js 6 and Node.js 8 runtimes.

2. Function Controller (FC) detects a new Function CR.

4. FC creates a ConfigMap with the lambda definition.

3. Based on the ConfigMap, FC creates a Kubernetes Job that triggers the creation of a lambda image.

4. The Job creates a Pod with the Docker image containing the lambda definition. It also pushes the image to a Docker registry.

5. FC monitors the Job status. When the image creation finishes successfully, FC creates a Service CR (KService) that points to the Pod with the image.

6. Knative Serving controller (KSC) detects the new KService and reads its definition.

7. KSC creates these resources:

    - Service Placeholder - a Kubernetes Service which has exactly the same name as the KService but [has no selectors](https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors) (does not point to any Pods). Its purpose is only to register the actual service name, such as `helloworld`, so it is unique. This service is exposed on port `80`.

    - Revision - a Kubernetes Service that KSC creates after any change in the KService. Each Revision is a different version of the KService. Revisions have selectors and point to specific Pods in a given Revision. Their names take the `{service-name}-{revision-number}` format, such as `helloworld-48thy` or `helloworld-vge8m`.

    - Virtual Service - a cluster-local Service that communicates only with resources within the cluster. This Virtual Service points to the Istio service mesh as a gateway to Service Revisions. The Virtual Service is registered for the name specified in the Service Placeholder, for example `helloworld.default.svc.cluster.local`.

        >**NOTE:** The **cluster-local** label in the KService instructs the KSC that it should not create an additional public Virtual Service.  

    - Route - a resource that redirects HTTP requests to specific Revisions.

    - Configuration - a resource that holds information on the desired Revision configuration.

    >**TIP:** For more details on all Knative Serving resources, read the official [Knative documentation](https://knative.dev/docs/serving/).

To sum up, the overall eventing communication with the lambda takes place through the Virtual Service that points to a given Revision. By default, the Virtual Service is set to the latest Revision but that can be modified to distribute traffic between Revisions if there is a high number of incoming events.

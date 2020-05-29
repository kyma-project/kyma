---
title: Architecture
---

Serverless relies on [Knative Serving](https://knative.dev/docs/serving/) for deploying and managing Functions and [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) for creating Docker images. See how these and other resources process a Function within a Kyma cluster:

![Serverless architecture](./assets/serverless-architecture.svg)

> **CAUTION:** Serverless imposes some requirements on the setup of Namespaces. If you create a new Namespace, do not disable sidecar injection in it as Serverless requires Istio for other resources to communicate with Functions correctly. Also, if you apply custom [LimitRanges](/root/kyma/#details-resource-quotas) for a new Namespace, they must be higher than or equal to the [limits for building Jobs' resources](#configuration-serverless-chart).

1. Create a Function either through the UI or by applying a Function custom resource (CR). This CR contains the Function definition (business logic that you want to execute) and information on the environment on which it should run.

    >**NOTE:** Function Controller sets the Node.js 12 runtime by default.

2. Before the Function can be saved or updated, it is first updated and then verified by the [defaulting and validation webhooks](#supported-webhooks) respectively.

3. Function Controller (FC) detects the new, validated Function CR.

4. FC creates a ConfigMap with the Function definition.

5. Based on the ConfigMap, FC creates a Kubernetes Job that triggers the creation of a Function image.

6. The Job creates a Pod which builds the production Docker image based on the Function's definition. The Job then pushes this image to a Docker registry.

7. FC monitors the Job status. When the image creation finishes successfully, FC creates a Service CR (KService) that points to the Pod with the image.

8. Knative Serving Controller (KSC) detects the new KService and reads its definition.

9. KSC creates these resources:

    - Service Placeholder - a Kubernetes Service which has exactly the same name as the KService but [has no selectors](https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors) (does not point to any Pods). Its purpose is only to register the actual service name, such as `helloworld`, so it is unique. This service is exposed on port `80`.

    - Revision - a Kubernetes Service that KSC creates after any change in the KService. Each Revision is a different version of the KService. Revisions have selectors and point to specific Pods in a given Revision. Their names take the `{service-name}-{revision-number}` format, such as `helloworld-48thy` or `helloworld-vge8m`.

    - Virtual Service - a cluster-local Service that communicates only with resources within the cluster. This Virtual Service points to the Istio service mesh as a gateway to Service Revisions. The Virtual Service is registered for the name specified in the Service Placeholder, for example `helloworld.default.svc.cluster.local`.

        >**NOTE:** The **cluster-local** label in the KService instructs the KSC that it should not create an additional public Virtual Service.  

    - Route - a resource that redirects HTTP requests to specific Revisions.

    - Configuration - a resource that holds information on the desired Revision configuration.

    >**TIP:** For more details on all Knative Serving resources, read the official [Knative documentation](https://knative.dev/docs/serving/).

To sum up, the overall eventing communication with the Function takes place through the Virtual Service that points to a given Revision. By default, the Virtual Service is set to the latest Revision but that can be modified to distribute traffic between Revisions if there is a high number of incoming events.

---
title: Architecture
---

Serverless v2 replies on [Knative Serving](https://knative.dev/docs/serving/) for deploying and managing functions, and [Tekton](https://github.com/tektoncd/pipeline) as a pipeline for creating and running on-cluster processes. See how these and other resources process a lambda within a Kyma cluster:

![Serverless v2 architecture](./assets/serverless-v2-architecture.svg)

1. The user creates a lambda either through the UI or by applying a Function custom resource (CR). This CR contains the lambda definition (business logic that the user wants to execute) and information on the environment on which it should be run (Node 6 or Node 8).

2. Function Controller (FC) detects a new Function CR and reads its definition.

3. Based on the Function CR definition, FC creates a [TaskRun CR](https://github.com/tektoncd/pipeline/blob/master/docs/taskruns.md), the purpose of which is to create an image based on the defined lambda.

4. Tekton Controller (TC) detects the new TaskRun CR and reads its definition.

5. Based on the TaskRun CR definition, TC triggers a pipeline that creates an image with lambda definition (Dockerfile). It later pushes this image to the local registry.

6. FC monitors the TaskRun CR. When the image creation finishes successfully, FC creates a Knative Service CR (KService) that points to the image.

7. Knative Serving controller (KSC) detects the new KService and reads its definition.

8. KSC creates these resources:

    - Service Placeholder - a Kubernetes Service which has exactly the same name as the KService but [has no selectors](https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors) (does not point to any Pods). Its purpose is only to register the actual service name, such as `helloword-go`, so it is unique.

    - Service Revision - a Kubernetes service for which KSC creates separate Revisions after any change in the KService. These Service Revisions have selectors and point to specific Pods in a given Revision. Their names take the `{service-name}-{revision-number}` format, such as `helloword-go-48thy` or `helloword-go-vge8m`.

    - Virtual Service - cluster-local Service that communicates only with resources within the cluster. This Virtual Service points to the Istio service mesh as a gateway to Service Revisions. The Virtual Service is registered for the name specified in the Service Placeholder, for example `helloword-go.default.svc.cluster.local`.

    >**NOTE:** The **cluster-local** label in the KService instructs the KSC that it should not create an additional public Virtual Service.  

To sum up, the overall eventing communication with our lambda takes place through the Virtual Service that points to a given Service Revision. By default, the Virtual Service is set to the latest Revision but that can be modified to distribute traffic between Revisions if there is a high number of incoming events.

# Serverless Architecture

Serverless relies heavily on Kubernetes resources. It uses [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/), [Services](https://kubernetes.io/docs/concepts/services-networking/service/) and [HorizontalPodAutoscalers](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) to deploy and manage Functions, and [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) to create Docker images. See how these and other resources process a Function within a Kyma cluster:

![Serverless architecture](../../assets/svls-architecture.svg)

> [!WARNING]
> Serverless imposes some requirements on the setup of namespaces. For example, if you apply custom [LimitRanges](https://kubernetes.io/docs/concepts/policy/limit-range/) for a new namespace, they must be higher than or equal to the limits for building Jobs' resources.

1. Create a Function either through the UI or by applying a Function custom resource (CR). This CR contains the Function definition (business logic that you want to execute) and information on the environment on which it should run.

2. Before the Function can be saved or modified, it is first updated and then verified by the defaulting and validation webhooks respectively.

3. Function Controller (FC) detects the new, validated Function CR.

4. FC creates a ConfigMap with the Function definition.

5. Based on the ConfigMap, FC creates a Kubernetes Job that triggers the creation of a Function image.

6. The Job creates a Pod which builds the production Docker image based on the Function's definition. The Job then pushes this image to a Docker registry.

7. FC monitors the Job status. When the image creation finishes successfully, FC creates a Deployment that uses the newly built image.

8. FC creates a Service that points to the Deployment.

9. FC creates a HorizontalPodAutoscaler that automatically scales the number of Pods in the Deployment based on the observed CPU utilization.

10. FC waits for the Deployment to become ready.

---
title: Access Kyma Application Logs
---

Get insights into your applications (microservices, Functions...) by viewing the respective logs.

To check out real-time logs immediately, use the Kubernetes functionalities - either with `kubectl`, or in Kyma dashboard.

## Kubernetes Logs in Kyma Dashboard

You can view real-time logs in Kyma dashboard:

1. Open Kyma dashboard and select the namespace.
2. Access the Pod and select the container.
3. Click **View Logs**.

## Kubernetes Logs Using kubectl

Alternatively, to see real-time logs in your terminal, run the following command:

```bash
kubectl logs {POD_NAME} --namespace {NAMESPACE_NAME} --container {CONTAINER_NAME}
```

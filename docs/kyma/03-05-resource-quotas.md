---
title: Resource quotas
type: Details
---

[Resource quotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/) are a convenient way to manage the consumption of resources in a Kyma cluster. You can easily set resource quotas for every Namespace you create through the Console UI.

When you click **Create Namespace**, you can define:
  - Total Memory Quotas, which limit the overall memory consumption by the Namespace by creating a ResourceQuota object.
  - Limits per container, which limit the memory consumption for individual containers in the Namespace by creating [LimitRange](https://kubernetes.io/docs/concepts/policy/limit-range/) objects.

To manage existing resource quotas in a Namespace, select that Namespace in the **Namespaces** view of the Console and go to the **Resources** tab. This view allows you to edit or delete the existing limits.

>**TIP:** If you want to manage ResourceQuotas and LimitRanges directly from the terminal, follow the [Kubernetes guide](https://kubernetes.io/docs/tasks/administer-cluster/manage-resources/quota-memory-cpu-namespace/).

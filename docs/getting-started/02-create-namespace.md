---
title: Create a Namespace
type: Getting Started
---

Almost all operations in these guides are performed using [Namespace](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)-scoped resources, so let's start by creating a dedicated Namespace called `orders-service` in which you will you will deploy all further resources.

## Reference

This and all other guides demonstrate steps you can perform both through terminal and UI. Read about the [Console](/components/console) through which you can visually and [securely](/components/security/) administer Kyma functionalities and manage the basic Kubernetes resources.

## Steps

Follow these steps:

<div tabs name="setup-namespace" group="set-up-namespace">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Create the `orders-service` Namespace:

   ```bash
   kubectl create ns orders-service
   ```

2. Check if the Namespace was set up. The Namespace status phase should state `Active`:

   ```bash
   kubectl get ns orders-service -o=jsonpath="{.status.phase}"
   ```

  </details>
  <details>
  <summary label="ui">
  UI
  </summary>

1. [Log into](/root/kyma#installation-install-kyma-on-a-cluster-access-the-cluster) the Kyma Console UI.

2. After logging, select **Add new namespace** in the **Namespaces** view.

3. Enter `orders-service` in the **Name** field.

4. Select **Create** to confirm the changes.

   You will be redirected to the `orders-service` Namespace view.

  </details>
</div>

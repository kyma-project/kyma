---
title: Create a Namespace
type: Getting Started
---

Almost all operations in these guides are performed using [Namespace](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)-scoped resources, so let's start by creating a dedicated `orders-service` Namespace for them.

## Reference

This and all other guides demonstrate steps you can perform in both the terminal (kubectl) and the Console UI. Read about the [Console](/components/console) through which you can graphically and [securely](/components/security/) administer Kyma functionalities and manage the basic Kubernetes resources.

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

2. Check that the Namespace was set up. This is indicated by the Namespace status phase `Active`:

   ```bash
   kubectl get ns orders-service -o=jsonpath="{.status.phase}"
   ```

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. [Log into](/root/kyma/#installation-install-kyma-on-a-cluster-access-the-cluster) the Console UI.

2. After logging in, select **Add new namespace** in the **Namespaces** view.

3. Enter `orders-service` in the **Name** field.

4. Select **Create** to confirm the changes.

   You will be redirected to the `orders-service` Namespace view.

  </details>
</div>

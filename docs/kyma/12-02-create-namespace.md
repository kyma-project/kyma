---
title: Set up a Namespace
type: Getting Started
---

Almost all operations in these guide are performed using Namespace-scoped resources, so let's start by creating a Namespace called `orders-service`.

Follow these steps:

<div tabs name="setup-namespace" group="set-up-namespace">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Create the Namespace:

   ```bash
   kubectl create ns orders-service
   ```

2. Check if the Namespace was set up successfully. The Namespace status phase should state `Active`:

   ```bash
   kubectl get ns orders-service -o=jsonpath="{.status.phase}"
   ```

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. [Log into](https://kyma-project.io/docs/1.12/root/kyma#installation-install-kyma-on-a-cluster-access-the-cluster) the Kyma Console UI.

2. After logging, select **Add new namespace** in the **Namespaces** view.

3. Enter `orders-service` in the **Name** field.

4. Select **Create** to confirm the changes.

   You will be redirected to the `orders-service` Namespace view.

  </details>
</div>

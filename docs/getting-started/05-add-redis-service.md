---
title: Add the Redis service
type: Getting Started
---

This tutorial shows how you can provision a sample [Redis](https://redis.io/) service using an Addon configuration linking to an example in the GitHub repository.

## Steps

Follows these steps:

<div tabs name="steps" group="create-redis-service">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Provision an AddonsConfiguration CR with the Redis service:

   ```yaml
   cat <<EOF | kubectl apply -f  -
   apiVersion: addons.kyma-project.io/v1alpha1
   kind: AddonsConfiguration
   metadata:
     name: redis-addon
     namespace: orders-service
   spec:
     repositories:
     - url: https://github.com/kyma-project/addons/releases/download/0.12.0/index-testing.yaml
   EOF
   ```

2. Check if the AddonsConfiguration CR was created successfully. The CR phase should state `Ready`:

  ```bash
  kubectl get addonsconfigurations redis-addon -n orders-service -o=jsonpath="{.status.phase}"
  ```

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. Navigate to the `orders-service` Namespace overview by selecting it from the drop-down list in the top navigation panel.

2. Navigate to the **Configuration** section in the left navigation panel, go to the **Addons** view, and select **Add New Configuration**.

3. Once the new box opens up, enter `https://github.com/kyma-project/addons/releases/download/0.11.0/index-testing.yaml` in the **Urls** field. The Addon name is automatically generated.

4. Select **Add** to confirm changes.

5. Wait for the Addon to have the `READY` status.


    </details>
</div>

---
title: Add a Redis service
type: Getting Started
---

This tutorial shows how you can provision a sample [Redis](https://redis.io/) service using an Addon configuration linking to an example in the GitHub repository.

## Steps

Follows these steps:

<div tabs name="steps" group="add-addon">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Export these variables:

   ```bash
   export NAME={FUNCTION_NAME}
   export NAMESPACE={FUNCTION_NAMESPACE}
   ```

2. Provision an Addon CR with the Redis service:

   ```yaml
   cat <<EOF | kubectl apply -f  -
   apiVersion: addons.kyma-project.io/v1alpha1
   kind: AddonsConfiguration
   metadata:
     name: $NAME
     namespace: $NAMESPACE
   spec:
     reprocessRequest: 0
     repositories:
     - url: https://github.com/kyma-project/addons/releases/download/0.11.0/index-testing.yaml
   EOF
   ```

3. Check if the Addon CR was created successfully. The CR phase should state `Ready`:

   ```bash
   kubectl get addons $NAME -n $NAMESPACE -o=jsonpath="{.status.phase}"
   ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. Select a Namespace from the drop-down list in the top navigation panel where you want to provision the Redis service.

2. Go to the **Addons** view in the left navigation panel and select **Add New Configuration**.

3. Enter `https://github.com/kyma-project/addons/releases/download/0.11.0/index-testing.yaml` in the **Urls** field. The Addon name is automatically generated.

4. Select **Add** to confirm changes.

   You will see that the Addon has the `Ready` status.

    </details>
</div>

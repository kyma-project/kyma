---
title: Add the Redis service
type: Getting Started
---

Provision the external [Redis](https://redis.io/) service that will replace your microservice in-memory storage. To do so, you will use a sample addon configuration linking to an example in the GitHub repository.

## Reference

This guide demonstrates how [Service Catalog](/components/service-catalog/) works in Kyma. It enables easy access to services ([ServiceClass custom resources](https://svc-cat.io/docs/resources/#service-classes)) managed by [Service Brokers](/components/service-catalog/#overview-service-brokers) registered in Kyma. You can consume these services by creating their instances and binding them to your microservices and Functions.

## Steps

Follows these steps:

<div tabs name="steps" group="create-redis-service">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Provision an [AddonsConfiguration custom resource (CR)](/components/helm-broker/#custom-resource-addons-configuration) with the Redis service:

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

2. Check that the AddonsConfiguration CR was created. This is indicated by the status phase `Ready`:

  ```bash
  kubectl get addonsconfigurations redis-addon -n orders-service -o=jsonpath="{.status.phase}"
  ```

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. Navigate to the `orders-service` Namespace overview by selecting it from the drop-down list in the top navigation panel.

2. Go to **Configuration** > **Addons** in the left navigation panel and select **Add New Configuration**.

3. Once the new box opens up, change **Name** to `redis-addon`, and enter `https://github.com/kyma-project/addons/releases/download/0.11.0/index-testing.yaml` in the **Urls** field.

4. Select **Add** to confirm the changes.

5. Wait for the addon to have the status `READY`.

    </details>
</div>

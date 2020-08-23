---
title: Expose a Function
type: Getting Started
---

This tutorial shows how you can expose a Function to access it outside the cluster, through an HTTP proxy. To expose it, use an APIRule custom resource (CR) managed by the in-house API Gateway Controller. This controller reacts to an instance of the APIRule CR and, based on its details, it creates an Istio Virtual Service and Oathkeeper Access Rules that specify your permissions for the exposed Function.

When you complete this tutorial, you get a Function that:

- Is available under an unsecured endpoint (**handler** set to `noop` in the APIRule CR).
- Accepts `GET` and `POST` methods.

>**NOTE:** Learn also how to [secure the Function](/components/api-gateway#tutorials-expose-and-secure-a-service-deploy-expose-and-secure-the-sample-resources).

## Prerequisites

This tutorial is based on an existing Function. To create one, follow the [Create a Function](/components/serverless#tutorials-create-a-function) tutorial.

## Steps

Follows these steps:

<div tabs name="steps" group="expose-function">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Create an APIRule CR for the Function. It is exposed on port `80` that is the default port of the [Service](/components/serverless#architecture-architecture).

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: orders-function
  namespace: orders-service
spec:
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
  - path: /.*
    accessStrategies:
    - config: {}
      handler: noop
    methods: ["GET","POST"]
  service:
    host: orders-function
    name: orders-function
    port: 80
EOF  
```

2. Check if the API Rule was created and has the `OK` status:

```bash
kubectl get apirules orders-function -n orders-service -o=jsonpath='{.status.APIRuleStatus.code}'
```

3. Access the Function's external address:

   ```bash
   curl https://orders-function.{CLUSTER_DOMAIN}
   ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. Navigate to the `orders-service` Namespace view in the Console UI from the drop-down list in the top navigation panel.

2. Go to the **API Rules** view under the **Configuration** section in the left navigation panel and select **Create API Rule**.

3. In the **General settings** section:

    - Enter `orders-function` as the API Rule's **Name**.

    >**NOTE:** The APIRule CR can have a different name than the Function, but it is recommended that all related resources share common names.

    - Enter `orders-function` as **Hostname** to indicate the host on which you want to expose your Function.

    - Select `orders-function` as the **Service** that indicates the Function you want to expose.

4. In the **Access strategies** section, leave the default settings, with `GET` and `POST` methods and the `noop` handler selected.

5. Select **Create** to confirm the changes.

    The message appears on the screen confirming the changes were saved.

6. Once the pop-up box closes, check if you can access the Function by selecting the HTTPS link under the **Host** column of the new `orders-function` API Rule.

    </details>
</div>

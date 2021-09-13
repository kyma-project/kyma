---
title: Expose the microservice
---

We have the Service created. Let's now expose it outside the cluster.

## Expose the microservice

To expose our microservice, we must create an [APIRule CR](../05-technical-reference/06-custom-resources/apix-01-apirule.md) for it, just like when we [exposed our Function](04-expose-function.md).

<div tabs name="Expose the microservice" group="deploy-microservice">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: orders-service
  namespace: default
  labels:
#    app: orders-service
#    example: orders-service
spec:
  service:
    host: orders-service
    name: orders-service
    port: 80
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
    - path: /.*
      methods: ["GET","POST"]
      accessStrategies:
        - handler: noop
      mutators: []
EOF
```

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. In your Services's view, click on **Expose Service +**.
2. Provide the **Name** (`hello-world`) and **Hostname** (`hello-world`) and click **Create**.

> **NOTE:** Alternatively, from the left navigation go to **APIRules**, click on **Create apirules +**, and continue with step 2, selecting the appropriate **Service** from the dropdown menu.
  </details>
</div>

## Verify the microservice exposure

Now let's check that the microservice has been exposed successfully.

<div tabs name="Verify microservice exposure" group="deploy-microservice">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
curl https://orders-service.$CLUSTER_DOMAIN/orders
```

The operation was successful if the command returns the (possibly empty `[]`) list of orders.

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. From your Services's view, get the APIRule's **Hostname**.

   > **NOTE:** Alternatively, from the left navigation go to **APIRules** and get the **Host** URL from there.

2. Paste this **Hostname** in your browser and add the `/orders` suffix to the end of it, like this: `{HOSTNAME}/orders`. Open it.

The operation was successful if the page shows the (possibly empty `[]`) list of orders.
  </details>
</div>




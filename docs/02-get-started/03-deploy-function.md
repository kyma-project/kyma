---
title: Deploy a Function
---

Now that you've installed Kyma, let's deploy your first Function. We'll call it `hello-world`.

## Create and apply your function

First, let's create the Function and apply it.

<div tabs name="Deploy a Function" group="deploy-function">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
kyma init function --name hello-world
kyma apply function
```

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. In Kyma Dashboard, go to the `default` Namespace.
2. Go to **Functions**.
3. Click on **Create Function +**.
4. Name the Function `hello-world` and click **Create**.
  </details>
</div>


## Verify the Function deployment

Now let's make sure that the Function has been deployed successfully. 

<div tabs name="Verify the Function deployment" group="deploy-function">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
kubectl get functions hello-world 
```

The operation was successful if the statuses for **CONFIGURED**, **BUILT**, and **RUNNING** are `True`.


  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

The operation was successful if the Function **Status** changed from `DEPLOYING` to `RUNNING`.

> **NOTE:** You might need to wait a few seconds for the status to change.
  </details>
</div>



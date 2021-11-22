---
title: Quick install
---

To get started with Kyma, let's quickly install it first.

## Install Kyma

To install Kyma on a local k3d cluster, run:

```bash
kyma provision k3d
kyma deploy
```

When asked whether to install the Kyma certificate, confirm.

> **NOTE:** Check out [more installation options for Kyma](../04-operation-guides/operations/02-install-kyma.md).

### Verify the installation

Now let's verify that the installation was successful. Run:

```bash
kubectl get deployments -n kyma-system
```

The installation succeeded if all the Deployments returned are in status `READY`.

### Export your cluster domain

For convenience, expose the domain of the cluster as an environment variable now. 
We will use it later in the guides. 

```bash
export CLUSTER_DOMAIN={YOUR_CLUSTER_DOMAIN}
```

> **NOTE:** For local installation, the cluster domain is `local.kyma.dev`.

## Connect to Kyma Dashboard

To manage Kyma via GUI, connect it to Kyma Dashboard. 

   To start the Dashboard, run:

    ```bash
    kyma dashboard
    ```
   This command takes you to your Kyma Dashboard under [`http://localhost:3001/`](http://localhost:3001/).

## Check the list of Deployments via Dashboard

Now let's check the list of deployments using the Dashboard.

1. Navigate to **Namespaces**.
2. Click on the `kyma-system` Namespace.
    > **NOTE:** The system Namespaces are hidden by default. 
    > To see `kyma-system` and other hidden Namespaces, go to your Dashboard profile, choose **Preferences** > **Clusters**, and activate the **Show Hidden Namespaces** toggle.
3. Go to **Workloads** > **Deployments**.

This gives you the same list of deployments as you got earlier via `kubectl`, just in a nicer visual packaging. 
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

1. To start the Dashboard, run:

    ```bash
    kyma dashboard
    ```
    This command opens the Dashboard under [`http://localhost:3001/`](http://localhost:3001/).
    
2. Click the button to add your cluster to the Dashboard. 
3. [Get your `kubeconfig` file](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/). Paste it into a text editor and replace the `0.0.0.0` part in the **cluster.server** value for k3d with `host.docker.internal`. <!-- //TODO: Remove the info about replacing the **cluster.server** value when this gets fixed. -->
4. Upload it into the Dashboard as prompted.

This takes you to your Kyma Dashboard.

<!--
//TODO: Finish when the following get fixed: Replace the Docker command in step 1 with `kyma dashboard --local` and the address in step 2 with `https://dashboard.kyma-project/`.
-->

## Check the list of Deployments via Dashboard

Now let's check the list of deployments using the Dashboard.

1. Navigate to **Namespaces**.
2. Click on the `kyma-system` Namespace.
    > **NOTE:** The system Namespaces are hidden by default. 
    > To see `kyma-system` and other hidden Namespaces, go to your Dashboard profile, choose **Preferences** > **Clusters**, and activate the **Show Hidden Namespaces** toggle.
3. Go to **Workloads** > **Deployments**.

This gives you the same list of deployments as you got earlier via `kubectl`, just in a nicer visual packaging. 
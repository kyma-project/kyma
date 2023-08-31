---
title: Quick install
---

To get started with Kyma, let's quickly install it first.

The Kyma project is currently in the transition phase from classic to modular Kyma. You can either install classic Kyma with its components, or available modules. To see the list of Kyma modules, go to [Overview](../01-overview/README.md). To learn how to install Kyma with a module, go to [Install, uninstall and upgrade Kyma with a module](08-install-uninstall-upgrade-kyma-module.md).

> **CAUTION:** Components transformed into modules aren't installed as part of preconfigured classic Kyma.

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

## Open Kyma Dashboard

To manage Kyma via GUI, open Kyma Dashboard:

```bash
kyma dashboard
```
This command takes you to your Kyma Dashboard under [`http://localhost:3001/`](http://localhost:3001/).

## Check the list of Deployments via Dashboard

Now let's check the list of deployments using the Dashboard.

1. Navigate to **Namespaces**.
2. Click on the `kyma-system` Namespace.
    > **NOTE:** The system Namespaces are hidden by default. 
    > To see `kyma-system` and other hidden Namespaces, go to your Dashboard profile in the top-right corner, choose **Preferences** > **Clusters**, and activate the **Show Hidden Namespaces** toggle.
3. Go to **Workloads** > **Deployments**.

This gives you the same list of deployments as you got earlier via `kubectl`, just in a nicer visual packaging. 
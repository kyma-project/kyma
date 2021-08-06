---
title: Provision a cluster
type: Tasks
---
<!-- this is surely mentioned in Basic Tasks/Get Started -->

To provision a k3d cluster, run:

```
kyma provision k3d
```
If you want to define the name of your k3d cluster and pass arguments to the Kubernetes API server (for example, to log to stderr), run:

```
kyma provision k3d --name='custom_name' --server-args='--alsologtostderr'
```
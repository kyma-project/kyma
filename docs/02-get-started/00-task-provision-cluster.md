---
title: Provision a cluster
type: Tasks
---
<!-- this is surely mentioned in Basic Tasks/Get Started -->

To provision a k3s cluster, run:

```
kyma alpha provision k3s 
```
If you want to define the name of your k3s cluster and pass arguments to the Kubernetes API server (for example, to log to stderr), run:

```
kyma alpha provision k3s --name='custom_name' --server-args='--alsologtostderr'
```
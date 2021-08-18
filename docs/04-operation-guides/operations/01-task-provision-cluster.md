---
title: Provision a cluster
---

To provision a k3d cluster, run:

```bash
kyma provision k3d
```

If you want to define the name of your k3d cluster and pass arguments to the Kubernetes API server (for example, to log to stderr), run:

```bash
kyma provision k3d --name='{CUSTOM_NAME}' --server-args='--alsologtostderr'
```
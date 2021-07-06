---
title: Error for kubectl port forwarding
---

## Condition

When executing `kubectl port-forward` I get the following error:

```bash
Unable to listen on port ... bind: address already in use
```

## Cause

Port forwarding failed because the local port is already reserved by another process running on your machine.

## Remedy

There are several ways to fix this:

* Kill the other process listening on the same port.
* Choose another (unused) local port.
* Let `kubectl` choose a random unused local port for you (see the [Kubernetes port forwarding documentation](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/#let-kubectl-choose-local-port)).

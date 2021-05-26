---
title: Issues with port forwarding
type: Troubleshooting
---

`kubectl port-forward` can fail with an error message `unable to listen on any of the requested ports` if the port is already reserved by another process running locally. To fix it, you can either kill the other process or change the local port to an unused one. You could also ask `kubectl` to choose a random unused local port for you: https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/#let-kubectl-choose-local-port.

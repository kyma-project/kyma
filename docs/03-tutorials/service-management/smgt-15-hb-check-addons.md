---
title: Check addons
---

Checker is a tool that validates addons in the [`addons`](https://github.com/kyma-project/addons) repository on every pull request. It checks whether all [required](./smgt-11-hb-create-addons.md) fields are set in your addons.

Checker also triggers the [`helm lint`](https://v2.helm.sh/docs/helm/#helm-lint) command using Helm CLI 2.16.1, which checks your addons' charts. Run Checker locally to test if your addons are valid:

```
go run components/helm-broker/cmd/checker/main.go {PATH_TO_YOUR_ADDONS}
```

If any addon does not meet the requirements, Helm Broker does not expose it as a Service Class. This situation is displayed in logs. To check logs from Helm Broker, run these commands:

```
export HELM_BROKER_POD_NAME=kubectl get pod -n kyma-system -l app=helm-broker
kubectl logs -n kyma-system $HELM_BROKER_POD_NAME helm-broker
```

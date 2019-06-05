---
title: Bundles validation
type: Details
---

The Checker is a tool that validates bundles in the [`bundles`](https://github.com/kyma-project/bundles) repository on every pull request. It checks whether all [required](#details-create-a-bundle) fields are set in your bundles.

The Checker also triggers the [helm lint](https://helm.sh/docs/helm/#helm-lint) command using helm CLI in 2.8.2 version, which checks your bundles' charts.

### Run the Checker locally

Run the Checker locally to test if your bundles are valid:
```
go run components/helm-broker/cmd/checker/main.go {PATH_TO_YOUR_BUNDLES}
```

## Troubleshooting

If any bundle does not meet the requirements, the Helm Broker does not expose it as a Service Class. This situation is displayed in logs.

To check logs from the Helm Broker, run these commands:

```
export HELM_BROKER_POD_NAME=kubectl get pod -n kyma-system -l app=helm-broker
kubectl logs -n kyma-system $HELM_BROKER_POD_NAME helm-broker
```

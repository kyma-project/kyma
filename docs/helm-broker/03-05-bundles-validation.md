---
title: Bundles validation
type: Details
---

The Checker tool is used in the [bundles](https://github.com/kyma-project/bundles) repo to validate a bundles on each pull request if all required fields are set. The requirements are described [here](03-01-create-bundles.md).

It's also triggers the [helm lint](https://helm.sh/docs/helm/#helm-lint) command using the `helm` in version `2.8.2`, which checks the bundle's chart.

### Run locally

You can run the Checker locally to test if your bundles are valid. To run it locally use the following command:
```
go run components/helm-broker/cmd/checker/main.go {PATH_TO_YOUR_BUNDLES}
```

## Troubleshooting

If some bundle does not meet the requirements, the Helm Broker won't expose it as a Service Class and it will put an information about this in the logs.

To check logs from the Helm Broker, execute that commands:

```
export HELM_BROKER_POD_NAME=kubectl get pod -n kyma-system -l app=helm-broker
kubectl logs -n kyma-system $HELM_BROKER_POD_NAME helm-broker
```

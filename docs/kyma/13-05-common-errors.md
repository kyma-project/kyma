---
title: Common installation errors
type: Troubleshooting
---

## Istio-related error

In some cases, the logs of Kyma installer may show this error, which seemingly indicates problems with Istio:

```
Step error:  Details: Helm install error: rpc error: code = Unknown desc = validation failed: [unable to recognize "": no matches for kind "DestinationRule" in version "networking.istio.io/v1alpha3", unable to recognize "": no matches for kind "DestinationRule" in version "networking.istio.io/v1alpha3", unable to recognize "": no matches for kind "attributemanifest" in version "config.istio.io/v1alpha2"
```

As Istio is the first sizeable component handled by the Installer, sometimes not all of the required CRDs are created before Installer proceeds to the next component. This situation doesn't cause the installation to fail.
Instead, the Istio installation step repeats and gets more time for setup and the error message is logged.

## Job failed: DeadlineExceeded error

The `Job failed: DeadlineExceeded` error indicates that a job object didn't finish in a set time leading to a timeout. Frequently this error is followed by a message that indicates the release which failed to install: `Helm install error: rpc error: code = Unknown desc = a release named core already exists`.

As this error is caused by a timeout, restart the installation.

If the problem repeats, find the job that causes the error and reach out to the ["installation"](https://kyma-community.slack.com/messages/CD2HJ0E78) Slack channel or create a [GitHub issue](https://github.com/kyma-project/kyma/issues). Follow these steps to identify the failing job:

1. Get the installed Helm releases which correspond to components:
  ```
  helm ls --tls
  ```
  A high number of revisions may suggest that a component was reinstalled several times. If a release has the status different to Deployed, the component wasn't installed.

2. Get component details:
  ```
  helm status {RELEASE_NAME} --tls
  ```
  Pods with not all containers in READY state might be the cause of the error.

3. Get the deployed jobs:
  ```
  kubectl get jobs --all-namespaces
  ```
  Jobs that are not completed might be the cause of the error.

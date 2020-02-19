---
title: Common installation errors
type: Troubleshooting
---

## Istio-related error

In some cases, the logs of the Kyma Installer may show this error, which seemingly indicates problems with Istio:

```
Step error:  Details: Helm install error: rpc error: code = Unknown desc = validation failed: [unable to recognize "": no matches for kind "DestinationRule" in version "networking.istio.io/v1alpha3", unable to recognize "": no matches for kind "DestinationRule" in version "networking.istio.io/v1alpha3", unable to recognize "": no matches for kind "attributemanifest" in version "config.istio.io/v1alpha2"
```

As Istio is the first sizeable component handled by the Kyma Installer, sometimes not all of the required CRDs are created before the Installer proceeds to the next component. This situation doesn't cause the installation to fail.
Instead, the Istio installation step repeats and gets more time for setup. The error message is logged regardless of that.

## Job failed: DeadlineExceeded error

The `Job failed: DeadlineExceeded` error indicates that a job object didn't finish in a set time leading to a time-out. Frequently this error is followed by a message that indicates the release which failed to install: `Helm install error: rpc error: code = Unknown desc = a release named core already exists`.

As this error is caused by a time-out, restart the installation.

If the problem repeats, find the job that causes the error and reach out to the **#installation** [Slack channel](http://slack.kyma-project.io/) or create a [GitHub issue](https://github.com/kyma-project/kyma/issues).

Follow these steps to identify the failing job:

1. Get the installed Helm releases which correspond to components:
  ```
  helm ls --tls
  ```
  A high number of revisions may suggest that a component was reinstalled several times. If a release has the status different to `Deployed`, the component wasn't installed.

2. Get component details:
  ```
  helm status {RELEASE_NAME} --tls
  ```
  Pods with not all containers in `READY` state can cause the error.

3. Get the deployed jobs:
  ```
  kubectl get jobs --all-namespaces
  ```
  Jobs that are not completed can cause the error.

## Installation fails without an apparent reason

If the installation fails and the feedback you get from the console output isn't sufficient to identify the root cause of the errors, use the `helm history` command to inspect errors that were logged for every revision of a given Helm release.

To list all of the available Helm releases, run:
```
helm list --tls
```
To inspect a release and its logged errors, run:
```
helm history {RELEASE_NAME} --tls
```

>**NOTE:** Names of Helm releases correspond to names of Kyma components.

## The server could not find the requested resource

During the installation process you may encounter `the server could not find the requested resource` error with misspelled CRD name:
```
Details: Helm install error: rpc error: code = Unknown desc = release compass failed: the server could not find the requested resource (post gatewaies.networking.istio.io)
```
Tiller in older versions is preparing plural names using set of rules, instead of reading them from the CRD. This method is not always producing the proper word. For example, `gatewaies` instead of `gateways`.

To resolve this error, upgrade Tiller. Run:
```
kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/{YOUR_KYMA_VERSION}/installation/resources/tiller.yaml
```

---
title: Common installation errors
type: Troubleshooting
---

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

During the installation process, you may encounter the `server could not find the requested resource` error that reports a misspelled CRD name:
```
Details: Helm install error: rpc error: code = Unknown desc = release compass failed: the server could not find the requested resource (post gatewaies.networking.istio.io)
```
Tiller in older versions prepares names plurals using a set of rules, instead of reading them from the CRD. This method does not always produce the correct plural form. For example, for `gateway` it produces `gatewaies` instead of `gateways`.

To resolve this error, upgrade Tiller. Run:
```
kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/{YOUR_KYMA_VERSION}/installation/resources/tiller.yaml
```

## Maximum number of retries reached

Kyma Installer retries installing a failed release a given number of times (default is 5). It stops the installation when it reaches the limit and returns this message: `Max number of retries reached during step {STEP_NAME}`. To start it again, check Kyma Installer logs to find out why this step failed, fix the issue, and follow the [update process](#installation-update-kyma).

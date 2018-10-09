# Stability Checker

## Overview
Purpose of the Stability Checker is to check if cluster is stable. To ensure that, Stability Checker is installed in the cluster and execute testing script in a loop. 
Notifications with test executions summary are send on a Slack channel.

## Installation

You can install the Stability Checker on the Kyma cluster as a helm chart. Find the chart definition in the `deploy/chart` directory.
1. Configure kubectl and helm to point your cluster by specifying **KUBECONFIG** environment variable. 
```
export KUBECONFIG="/path/to/kubeconfig"
```
1. Provision volume with testing script, which then will be used by the Stability Checker. For that purpose you can use 
script `local/provision_volume.sh`. Script copies all files placed in `local/input' directory to PV.

2. Install stability checker as a helm chart and provide proper configuration:

```
helm install deploy/chart/stability-checker \
  --set slackClientWebhookUrl="" \
  ...
  --namespace="kyma-system" \
  --name="stability-checker"

```

Below you can find configuration options:

 | Name | Default value | Description |
 |------|---------------|-------------|
containerRegistry.path |eu.gcr.io/kyma-project| Address of Docker registry where Stability Checker image can be found.
image.tag |8b5d53d3| Tag of Stability Checker docker image.
image.pullPolicy |IfNotPresent| K8s image pull policy.
storage.claimName | Name of PVC which is attached to Stability Checker pod. Volume is visible in the pod under "/data" path. 
pathToTestingScript |/data/input/testing.sh| Full path to testing script. Because script is delivered inside PV, it has to start with "/data".
slackClientWebhookUrl |need-to-be-provided| Slack client webhook URL.
slackClientChannelId |need-to-be-provided| Slack channel ID, starts with `#`.
slackClientToken |123-need-to-be-provided| Slack client token.
testThrottle | 5m | Period between test executions. Purpose of this time is to give K8s time to clean up all resources after previous test.
testResultWindowTime | 6h | Notifications will be sent after this time and contains test executions summary for this period. 
stats.enabled | false | If true, output from test executions is analyzed to find statistics for every test. 
Detailed information about how many times every test failed and succeeded will be enclosed to the slack notification. 
Detecting test result is done by regular expressions defined in `stats.failingTestRegexp` and `stats.successfulTestRegexp`
stats.failingTestRegexp |TBD| Regular expression which indicate that test has failed. Has to contain one capturing group which identifies test name.
stats.successfulTestRegexp |TBD|  Regular expression which indicate that test has passed. Has to contain one capturing group which identifies test name.
service.type |NodePort| Stability Checker service type
service.externalPort |80| Stability Checker service external port
service.internalPort |8080| Stability Checker service internal port

> **NOTE:** You must install the chart after running the core tests, to avoid running the same tests in parallel.
Following values can be specified for chart:

## Development
Use the following helpers for the local development:
- `./local_minikube_build.sh` which builds the Stability Checker Docker image on a  Minikube registry.
- `./local/provision_volume.sh` which provisions a PersistentVolumeClaim (PVC) with testing scripts.
- `./local/charts/dummy` chart which contains simple and fast tests. To install it, execute the following command:
```
    helm install ./dummy --name dummy --namespace=kyma-system
```
- `./local_helm_install.sh` which installs the Stability Checker Helm chart with predefined values. 
The testing script points to the `testing-dummy.sh` which is a simplified version of `testing-kyma.sh`. The`dummy` chart is used in the `testing-dummy.sh` to speed up testing.
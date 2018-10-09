# Stability Checker

## Overview
Purpose of the Stability Checker is to ensure that cluster is stable. To ensure that, we execute in a loop testing script. 
The Stability Checker runs the testing script in a loop  and reports the results on a Slack channel.

## Installation

You can install the Stability Checker on the Kyma cluster as a helm chart. Find the chart definition in the `deploy/chart` directory.
1. Configure kubectl and helm to point your cluster:
```
export KUBECONFIG="/path/to/kubeconfig"
```
1. Provision volume with testing script, which then will be used by the Stability Checker. For that purpose you can use 
script `local/provision_volume.sh`. Script copies all files placed in `local/input' directory.

2. Install stability checker with following command:

```
export KUBECONFIG=""
helm install deploy/chart/stability-checker \
  --set slackClientWebhookUrl="TBD" \
  --set slackClientChannelId="TBD" \
  --set slackClientToken="TBD" \
  --set testThrottle="1m" \
  --set testResultWindowTime="10m" \
  --set stats.enabled="true" \
  --set stats.failingTestRegexp="FAILED: ([0-9A-Za-z_-]+)" \
  --set stats.successfulTestRegexp="PASSED: ([0-9A-Za-z_-]+)" \
  --set pathToTestingScript="/data/input/testing-dummy.sh" \
  --namespace="kyma-system" \
  --name="stability-checker"

```

> **NOTE:** You must install the chart after running the core tests, to avoid running the same tests in parallel.
Following values can be specified for chart:

 | Name | Default value | Description |
 |------|---------------|-------------|
containerRegistry.path |eu.gcr.io/kyma-project| Address of Docker registry where for Stability Checker image
image.tag |8b5d53d3| Tag of Stability Checker docker image
image.pullPolicy |IfNotPresent| K8s image pull policy
storage.claimName | Name of PVC which is attached to Stability Checker pod. Volume is mounted to "/data" |
pathToTestingScript |/data/input/testing.sh| Full path to testing script. It has to 
slackClientWebhookUrl |need-to-be-provided| Slack client webhook URL
slackClientChannelId |need-to-be-provided| Slack channel ID, should start with `#`
slackClientToken |123-need-to-be-provided| Slack client token.
testThrottle | 5m | Period between test executions. 
testResultWindowTime | 6h | Notifications will be sent after this time and contains test execution summary from this period. 
stats.enabled | false | If gather statistics from test executions for every single test and then sent this information in notification. 
stats.failingTestRegexp |TBD| Regular expression which indicate that test has failed. Has to contain one capturing group which identifies test name.
stats.successfulTestRegexp |TBD|  Regular expression which indicate that test has passed. Has to contain one capturing group which identifies test name.
service.type |NodePort|
service.externalPort |80|
service.internalPort |8080|



## Usage

Stability Checker does not contain testing scripts. The chart value `.Values.pathToTestingScript` defines which script the system runs.
Ensure you have the file available in a persistent volume defined as `.Values.storage.claimName`, which is mounted as a `data` directory in the Stability Checker Pod.

To simulate the process of providing scripts, see the `local/provision_volume.sh` script, which populates the volume with files from the `local/input` directory.

### Deliver the chart

Download the gzipped chart from:

`https://github.com/kyma-project/stability-checker/raw/{branchName}/deploy/chart/stability-checker-0.1.0.tgz`

As another option, you can run the following:

```helm install https://github.com/kyma-project/stability-checker/raw/{branchName}/deploy/chart/stability-checker-0.1.0.tgz```

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
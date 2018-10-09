# Stability Checker

## Overview
Purpose of the Stability Checker is to check if a cluster is stable. To ensure that, Stability Checker is installed in the cluster and execute testing script in a loop. 
Notifications with test executions summary are sent on a Slack channel.

## Installation

You can install the Stability Checker on the Kyma cluster as a helm chart. Find the chart definition in the `deploy/chart` directory.
1. Configure kubectl and helm to point your cluster by specifying **KUBECONFIG** environment variable. 
```
export KUBECONFIG="/path/to/kubeconfig"
```
1. Provision volume with a testing script, which then will be used by the Stability Checker. For that purpose, you can use 
script `local/provision_volume.sh`. The script copies all files placed in `local/input` directory to PV.

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
storage.claimName |stability-test-scripts-pvc| Name of PVC which is attached to Stability Checker pod. Volume is visible in the pod under `/data` path. 
pathToTestingScript |/data/input/testing.sh| Full path to the testing script. Because script is delivered inside PV, it has to start with `/data`.
slackClientWebhookUrl |-| Slack client webhook URL.
slackClientChannelId |-| Slack channel ID, starts with `#`.
slackClientToken |-| Slack client token.
testThrottle | 5m | Period between test executions. Purpose of this parameter is to give K8s time to clean up all resources after the previous test execution.
testResultWindowTime | 6h | Notifications will be sent after this time and contains test executions summary for this period. 
stats.enabled | false | If true, an output from test executions is analyzed to find statistics for every specific test. Detailed information about how many times every test failed and succeeded will be enclosed to the slack notification. Detecting test result is done by regular expressions defined in `stats.failingTestRegexp` and `stats.successfulTestRegexp`.
stats.failingTestRegexp |-| Regular expression which indicates that test has failed. Has to contain one capturing group which identifies test name.
stats.successfulTestRegexp |-|  Regular expression which indicates that the test has passed. Has to contain one capturing group which identifies test name.


> **NOTE:** You must install the chart after running the core tests, to avoid running the same tests in parallel.
Following values can be specified for chart:

## Development
Use the following helpers for the local development:
- `./local_minikube_build.sh` which builds the Stability Checker Docker image on a  Minikube registry.
- `./local/provision_volume.sh` which provisions a PersistentVolumeClaim (PVC) with testing scripts.
- `./local/charts/dummy` the chart which contains simple and fast tests. To install it, execute the following command:
```
    helm install ./dummy --name dummy --namespace=kyma-system
```
- `./local_helm_install.sh` which installs the Stability Checker Helm chart with predefined values. 
The testing script points to the `testing-dummy.sh` which is a simplified version of `testing-kyma.sh`. The`dummy` chart is used in the `testing-dummy.sh` to speed up testing.
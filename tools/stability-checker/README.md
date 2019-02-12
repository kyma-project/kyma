# Stability Checker

## Overview
The purpose of the Stability Checker is to check if a cluster is stable. 
To ensure the cluster's stability, install the Stability Checker in the cluster. 
It runs a testing script in a loop and sends notifications with the test executions summary on a Slack channel.

## Installation
Install the Stability Checker on the Kyma cluster as a Helm chart. Find the chart definition in the `deploy/chart` directory.
1. Configure kubectl and helm to point to your cluster by specifying the **KUBECONFIG** environment variable. 
```
export KUBECONFIG="{path to kubeconfig}"
```

> **NOTE:** Ensure that the kubeconfig allows you to update the cluster state by installing Helm charts. 
2. Provision volume with a testing script, which then will be used by the Stability Checker. You can use the `local/provision_volume.sh` script. 
The script copies all files placed in the `local/input` directory to PV.

3. The Stability Checker does not send logs to the Google Cloud Storage bucket by default.
To enable log forwarding, create the Secret with credentials to the google account and name it `sa-stability-fluentd-storage-writer`.
Additionally, set the **logsPersistence.enabled** parameter in the `deploy/chart/stability-checker/values.yaml` file to `true`.

4. Install the Stability Checker as a Helm chart and provide the proper configuration:

```
helm install deploy/chart/stability-checker \
  --set slackClientWebhookUrl="" \
  ...
  --namespace="kyma-system" \
  --name="stability-checker"

```

The configuration options are as follows:

 | Name | Default value | Description |
 |------|---------------|-------------|
**storage.claimName** |`stability-test-scripts-pvc`| Name of the Persistent Volume Claim (PVC) which is attached to the Stability Checker Pod. The volume is visible in the Pod under the `/data` path. 
**clusterName** |-| Name of the cluster on which the Stability Checker Pod is launched. The cluster name is attached to the tests summary report and sent to the given Slack channel.
**pathToTestingScript** |`/data/input/testing.sh`| Full path to the testing script. As the script is delivered inside the PVC, the path must start with `/data`.
**slackClientWebhookUrl** |-| Slack client webhook URL. For details, click **Customize Slack** in your Slack Workspace, choose the **Configure apps** button and  proceed with the configuration of the `Jenkins CI` application.
**slackClientChannelId** |-| Slack channel ID which starts with `#`.
**slackClientToken** |-| Slack client token. For details, click **Customize Slack** in your Slack Workspace, choose the **Configure apps** button and  proceed with the configuration of the `Jenkins CI` application.
**testThrottle** | `5m`| Period between test executions. The purpose of this parameter is to give Kubernetes some time to clean up all resources after the previous test execution.
**testResultWindowTime** | `6h` | Time period after which the Stability Checker sends notifications. Notifications contain a test executions summary for this period. 
**stats.enabled** | `false` | If set to `true`, an output from test executions is analyzed to find statistics for every specific test. Detailed information about the number of times every test failed and succeeded is enclosed to the Slack notification. Regular expressions defined in **stats.failingTestRegexp** and **stats.successfulTestRegexp** detect test results. You can configure these two parameters only if **stats.enabled** is set to `true`. 
**stats.failingTestRegexp** |-| Regular expression which indicates that the test has failed. It must contain one capturing group which identifies the test name.
**stats.successfulTestRegexp** |-|  Regular expression which indicates that the test has passed. It must contain one capturing group which identifies the test name.


> **NOTE:** You must install the chart after running the core tests, to avoid running the same tests in parallel.

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

# Load Test


## Overview

The purpose of the load test is to verify the execution of the Horizontal Pod Autoscaling of functions deployed with the help of Kubeless in **Kyma**. It runs a **golang** app which stress a function by making thousands of HTTP POST request in set period of time defined in an environment variable. At the end of the execution of the test a notification is send to a Slack channel with the final output of the load test.

## Installation

```bash
echo "Installing helm chart..."
helm install --set slackEndpoint="${SLACK_ENDPOINT}" \
             --set slackClientToken="${SLACK_CLIENT_TOKEN}" \
             --set slackClientChannelId="${SLACK_CLIENT_CHANNEL_ID}" \
             --set loadTestExecutionTimeout="$LT_TIMEOUT" \
             ${LT_FOLDER_CHART} \
             --namespace=kyma-system \
             --name=load-test
```

The configuration needed to execute the **load test** is as follows:

 | Name | Value | Description |
 |------|---------------|-------------|
**slackEndpoint** |`-`| A webhook slack url.
**slackClientToken** |`-`|  A token which will be part of the **slackEndpoint**.
**slackClientChannelId** |`#channelId`| ID of the Slack channel.
**loadTestExecutionTimeout** |`30`| Time to finish the test otherwise it will timeout its execution. All the related metrics to be sent to the Slack channel are collected after the timeout.

### Environment Variables

To run the **load test** either in a cluster or a local Minikube you will need to set the above parameters for the installation of the helm chart. However in a cluster it is advisable to have environment variables as it is shown above.

```bash
### Test config
# export SLACK_CLIENT_TOKEN='replace by the slack token'
# export SLACK_CLIENT_CHANNEL_ID='replace by the slack #channel'
# export SLACK_ENDPOINT='replace by the webhook http endpoint https://endpoint_here'
# export LT_FOLDER_CHART='HOME_KYMA_PROJECT/kyma/tools/load-test/deploy/chart/load-test'
# export LT_TIMEOUT=30
```

## Development

- **load-test/k8syaml**  contains all the Kubernetes resources needed to deploy the function.

- **main.go** All the logic of the **load test** can be found in this file. It can be built as follows:
 
 `GOOS=linux GOARCH=amd64 go build -o ./bin/app`(Mac)
 `CGO_ENABLED=0 go build -o ./bin/app`(Linux)
 
- **load-test/Dockerfile** is a file needed to build the Docker image. To build the image, run this command:

`docker build -t load-test .`

- **load-test/deploy/chart** contains the chart that installs the test code which stresses the function.

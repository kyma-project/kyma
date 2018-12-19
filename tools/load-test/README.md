# Load Test


## Overview

The load test verifies the execution of the Horizontal Pod Autoscaling of functions deployed using Kubeless in Kyma. It runs a Golang application which stresses a function by making thousands of HTTP POST requests in a set period of time defined in an environment variable. At the end of the test execution, a notification with the output is sent to a Slack channel.

## Installation

Run the following script to install the test:

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

To run the load test either on a cluster or on a local Minikube, you must set the Helm chart installation parameters. However, on a cluster it is advisable to have the environment variables set to default values.

These are the parameters you must configure to execute the load test:

 | Parameter | Value | Description |
 |------|---------------|-------------|
**slackEndpoint** |`-`| A webhook Slack URL.
**slackClientToken** |`-`|  A token which is a part of the **slackEndpoint**.
**slackClientChannelId** |`#channelId`| ID of the Slack channel.
**loadTestExecutionTimeout** |`30`| Time after which the test execution timeout occurs. All the related metrics to be sent to the Slack channel are collected after the timeout.

You can also define the values of these parameters while running this command:

```bash
### Test config
# export SLACK_CLIENT_TOKEN='replace by the slack token'
# export SLACK_CLIENT_CHANNEL_ID='replace by the slack #channel'
# export SLACK_ENDPOINT='replace by the webhook http endpoint https://endpoint_here'
# export LT_FOLDER_CHART='HOME_KYMA_PROJECT/kyma/tools/load-test/deploy/chart/load-test'
# export LT_TIMEOUT=30
```

## Development

These are the main components of the load test:

- **load-test/k8syaml**  contains all the Kubernetes resources needed to deploy the function.

- **load-test.go** contains all the logic of the load test. You can build it with these commands:
 
 ```
    GOOS=linux GOARCH=amd64 go build -o ./bin/app (Mac)
    CGO_ENABLED=0 go build -o ./bin/app (Linux)
 ``` 
 
- **load-test/Dockerfile** is a file needed to build the Docker image. To build the image, run this command:

```
    docker build -t load-test .
```

- **load-test/deploy/chart** contains the chart that installs the test code which stresses the function.

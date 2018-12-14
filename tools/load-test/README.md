# Load Test


## Overview

The purpose of the load test is to verify the right execution of the Horizonatl Pod Autoscaling in **Kyma**. It runs a **golang** app which stress a function by making thousands of http post request in set period of time defined in an environment variable. At the end of the execution of the test a notification is send to a Slack channel with the final output of the load test.

## Installation

```bash
echo "Installing helm chart..."
helm install --set slackEndpoint="${SLACK_ENDPOINT}" \
             --set slackClientToken="${SLACK_CLIENT_TOKEN}" \
             --set slackClientChannelId="${SLACK_CLIENT_CHANNEL_ID}" \
             --set loadTestExecutionTimeout="30" \
             ${LT_FOLDER}/deploy/chart/load-test \
             --namespace=kyma-system \
--name=load-test
```

The configuration needed in oder to execute the **load test**:

 | Name | Value | Description |
 |------|---------------|-------------|
**slackEndpoint** |`-`| A weebhook slack url.
**slackClientToken** |`-`|  A token which will be part of the **slackEndpoint**. 
**slackClientChannelId** |`#channelId`| ID of the Slach channel.
**loadTestExecutionTimeout** |`30`| time in which the test will timeout an finishing its execution. All the related metrics to be send to the Slack channel are collected after the timeout.

Run the **load test** either is a cluster or a local Minikube will need to set the above parameters for the install of the helm chart. However in a cluster it is advisable to have environment variables as it shown above.


## Development

- **load-test/k8syaml**  contains all the kubernetes resources needed to deploy the function.

- **main.go** All the logic of the **load test** can be found in this file. It can be build as a follow:
 
 `CGO_ENABLED=0 go build -o ./bin/app`
 
- **load-test/Dockerfile** Needed to build the docker image. The image can be build as a follow:

`docker build -t load-test .`

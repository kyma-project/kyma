# Load Test


## Overview

The purpose of the load test is to verify the right execution of the Horizonatl Pod Autoscaling in **Kyma**. It runs a **golang** app which stress a function by making thousands of http post request in set period of time defined in an environment variable. At the end of the execution of the test a notification is send to a Slack channel with the final output of the load test.

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

The configuration needed in oder to execute the **load test**:

 | Name | Value | Description |
 |------|---------------|-------------|
**slackEndpoint** |`-`| A webhook slack url.
**slackClientToken** |`-`|  A token which will be part of the **slackEndpoint**. 
**slackClientChannelId** |`#channelId`| ID of the Slach channel.
**loadTestExecutionTimeout** |`30`| time in which the test will timeout an finishing its execution. All the related metrics to be sent to the Slack channel are collected after the timeout.

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

- **load-test/k8syaml**  contains all the kubernetes resources needed to deploy the function.

- **main.go** All the logic of the **load test** can be found in this file. It can be built as a follows:
 
 `CGO_ENABLED=0 go build -o ./bin/app`
 
- **load-test/Dockerfile** Needed to build the docker image. The image can be built as a follows:

`docker build -t load-test .`

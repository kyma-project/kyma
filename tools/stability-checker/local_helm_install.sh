#!/usr/bin/env bash

helm install deploy/chart/stability-checker \
  --set containerRegistry.path="local" \
  --set image.tag="local" \
  --set clusterName="TBD" \
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

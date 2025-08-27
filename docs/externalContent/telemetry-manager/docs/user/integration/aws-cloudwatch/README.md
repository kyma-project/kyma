# Integrate Kyma with Amazon CloudWatch

## Overview

| Category|                       |
| - |-----------------------|
| Signal types | traces, logs, metrics |
| Backend type | third-party remote    |
| OTLP-native | no                    |

Learn how to use [Amazon CloudWatch](https://aws.amazon.com/cloudwatch) as backend for the Kyma Telemetry module.

Because CloudWatch doesn't support native OTLP ingestion for metrics, and OTLP support for logs and traces need the AWS-specific [segv4](https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-authenticating-requests.html) authentication, the Telemetry module must first ingest the signals into a custom OTel Collector based on the [Contrib distribution](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-OTLPSimplesetup.html) of the collector. Then, the custom collector converts the signals to the format required by CloudWatch and ingests them into CloudWatch.

![overview](../assets/cloudwatch.drawio.svg)

## Table of Content

- [Integrate Kyma with Amazon CloudWatch](#integrate-kyma-with-amazon-cloudwatch)
  - [Overview](#overview)
  - [Table of Content](#table-of-content)
  - [Prerequisites](#prerequisites)
  - [Prepare the Namespace](#prepare-the-namespace)
  - [Set Up AWS Credentials](#set-up-aws-credentials)
    - [Create AWS IAM User](#create-aws-iam-user)
    - [Create a Secret with AWS Credentials](#create-a-secret-with-aws-credentials)
  - [Deploy the Custom Collector](#deploy-the-custom-collector)
  - [Set Up Kyma Pipelines](#set-up-kyma-pipelines)
  - [Verify the Results](#verify-the-results)

## Prerequisites

- Kyma as the target deployment environment
- The [Telemetry module](https://kyma-project.io/#/telemetry-manager/user/README) is [added](https://kyma-project.io/#/02-get-started/01-quick-install)
- [Kubectl version that is within one minor version (older or newer) of `kube-apiserver`](https://kubernetes.io/releases/version-skew-policy/#kubectl)
- AWS account with permissions to create new users and security policies
- AWS CloudWatch configured with a [LogGroup and LogStream](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/Working-with-log-groups-and-streams.html)
- AWS CloudWatch configured with [Transaction Search](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-Transaction-Search.html) enabled

## Prepare the Namespace

1. Export your namespace as a variable with the following command:

    ```bash
    export K8S_NAMESPACE="aws"
    ```

1. If you haven't created a namespace yet, do it now:

    ```bash
    kubectl create namespace $K8S_NAMESPACE
    ```

## Set Up AWS Credentials

### Create AWS IAM User

1. In your AWS account, create an [IAM user](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users.html) and attach the **CloudWatchAgentServerPolicy** policy.
1. For the [IAM user](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users.html) you just created, create an access key for an application running outside AWS. Copy and save the **access key** and **secret access key**; you need them to [Create a Secret with AWS Credentials](#create-a-secret-with-aws-credentials).

### Create a Secret with AWS Credentials

To connect the AWS Distro to the AWS services, create a Secret containing the credentials of the created IAM user into the Kyma cluster. In the following command, replace `{ACCESS_KEY}` with your access key, `{SECRET_ACCESS_KEY}` with your Secret access key, and `{AWS_REGION}` with the AWS region you want to use:

   ```bash
   kubectl create secret generic aws-credentials -n $K8S_NAMESPACE --from-literal=AWS_ACCESS_KEY_ID={ACCESS_KEY} --from-literal=AWS_SECRET_ACCESS_KEY={SECRET_ACCESS_KEY} --from-literal=AWS_REGION={AWS_REGION}
   ```

## Deploy the Custom Collector

1. Export the Helm release name that you want to use. The release name must be unique for the chosen namespace. Be aware that all resources in the cluster will be prefixed with that name. Run the following command:

    ```bash
    export HELM_OTEL_AWS_RELEASE="aws"
    ```

1. Update your Helm installation with the required Helm repository:

    ```bash
    helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
    helm repo update
    ```

1. Deploy the custom collector using Helm by calling:

```sh
helm upgrade --install -n $K8S_NAMESPACE $HELM_OTEL_AWS_RELEASE open-telemetry/opentelemetry-collector -f https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/aws-cloudwatch/values.yaml
```
<!-- markdown-link-check-disable -->
The previous command uses the [values.yaml](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/aws-cloudwatch/values.yaml) provided in this `aws-cloudwatch` folder, which contains customized settings deviating from the default settings in the following ways:
<!-- markdown-link-check-enable -->

- Mount the values of Secret `aws-credentials` as environment variables
- Configure the OTel exporter for logs assuming a LogGroup `/logs/kyma` and LogStream `default`
- Configure the OTel exporter for traces
- Configure the AWSEMF exporter for metrics

## Set Up Kyma Pipelines

Use the Kyma Telemetry module to enable ingestion of the signals from your workloads:

1. Deploy a [LogPipeline](./../../02-logs.md):

    ```bash
    kubectl apply -f - <<EOF
    apiVersion: telemetry.kyma-project.io/v1alpha1
    kind: LogPipeline
    metadata:
      name: aws
    spec:
      input:
        application:
          enabled: true
      output:
        otlp:
          endpoint:
            value: http://$HELM_OTEL_AWS_RELEASE-opentelemetry-collector.$K8S_NAMESPACE:4317

   EOF
   ```

2. Deploy a [TracePipeline](./../../03-traces.md):

   ```bash
   kubectl apply -f - <<EOF
   apiVersion: telemetry.kyma-project.io/v1alpha1
   kind: TracePipeline
   metadata:
     name: aws
   spec:
     output:
       otlp:
         endpoint:
           value: http://$HELM_OTEL_AWS_RELEASE-opentelemetry-collector.$K8S_NAMESPACE.svc.cluster.local:4317
   EOF
   ```

3. Deploy a [MetricPipeline](./../../04-metrics.md):

   ```bash
   kubectl apply -f - <<EOF
   apiVersion: telemetry.kyma-project.io/v1alpha1
   kind: MetricPipeline
   metadata:
     name: awsh
   spec:
     input:
       runtime:
         enabled: true
     output:
       otlp:
         endpoint:
           value: http://$HELM_OTEL_AWS_RELEASE-opentelemetry-collector.$K8S_NAMESPACE.svc.cluster.local:4317
   EOF
   ```

## Verify the Results

Verify that the logs, traces, and metrics are exported to CloudWatch.

1. [Install the OpenTelemetry demo application](../opentelemetry-demo/README.md).
2. Go to `https://{AWS_REGION}.console.aws.amazon.com/cloudwatch`. Replace `{AWS_REGION}` with the region that you have chosen when [creating the Secret with AWS credentials](#create-a-secret-with-aws-credentials).
3. Verify that data is visible for all three signal types.

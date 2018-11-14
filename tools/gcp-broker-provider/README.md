# gcp-broker-provider

## Overview

This image provides tools and scripts to install GCP Service Broker in 

The gcp-broker-provider image has the following binaries installed:

* kubectl
* gcloud (cloud sdk)
* sc - https://github.com/kyma-incubator/k8s-service-catalog

For more details, see this [Dockerfile](Dockerfile).

## Prerequisites

To set up the project, use these tools:
* The latest version of [Docker](https://www.docker.com/)

## Usage

To build gcp-broker-provider locally, run this command:

```bash
docker build --no-cache -t gcp-broker-provider
```

Kubectl and SC version can be overridden using this build-args:
```bash
--build-arg KUBECTL_CLI_VERSION=${KUBECTL_CLI_VERSION}
--build-arg SC_CLI_VERSION=${SC_CLI_VERSION}
```

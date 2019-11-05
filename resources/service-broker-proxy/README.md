# Service Broker Proxy k8s

## Overview
This helm chart bootstraps the Service Broker Proxy for Kubernetes to connect Kyma with Service Manager.

## Prerequisites

Before provision the Service Broker Proxy credential to Service Manager should be updated.
Credential are located in `values.yaml` under two fields `sm.user` and `sm.password` and used by `smsecret.yaml` file.

The URL to Service Manager can be overridden. Default URL is located in `values.yaml` under the `config.sm.url` key.

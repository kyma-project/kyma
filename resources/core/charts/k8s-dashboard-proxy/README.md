# Kubernetes Dashboard proxy

## Overview

The Kubernetes Dashboard proxy is a transparent proxy for the Kubernetes Dashboard. Read the [Kubernetes Dashboard proxy](https://github.com/kyma-project/kyma/blob/master/components/k8s-dashboard-proxy/README.md) document to learn more about Kubernetes Dashboard proxy implementation in Kyma.

## Details

This chart installs [Keycloak proxy](https://github.com/keycloak/keycloak-gatekeeper) and **reverseproxy**, a thin layer Go application.

Configure the following options for each business requirement:

| Component                 | Configuration  | Description |
|---------------------------| --------------:| ----------: |
| **k8s-dashboard-proxy** |
| |`containerPort` | The port used by the proxy to listen on. |
| **reverseproxy** |
| |`host`| The upstream host. |
| |`port`| The upstream port. |
| |`secret_token_path`| The path of Kubernetes Service Account Secret. |
| |`k8s_dashboard_URL`| The Kubernetes Dashboard URL. |
| **keycloak** |
| |`image`| The Keycloak proxy Docker image. |

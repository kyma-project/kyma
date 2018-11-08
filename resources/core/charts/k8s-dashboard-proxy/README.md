# Kubernetes Dashboard Proxy

## Overview

The `Kubernetes Dashboard Proxy` is a transparent proxy for the `Kubernetes Dashboard`

Read the [Kubernetes Dashboard Proxy](https://github.com/kyma-project/kyma/blob/master/components/k8s-dashboard-proxy/README.md) document to learn more about Kubernetes Dashboard Proxy in Kyma.

This chart installs [Keycloak Proxy](https://github.com/keycloak/keycloak-gatekeeper) and `reverseproxy`, a thin layer Go application as a reverse proxy.

Configure these options for each business requirement:

| Component                 | Configuration  | Description |
|---------------------------| --------------:| ----------: |
| **k8s-dashboard-proxy** |
| |`containerPort` | The proxy listen port. |
| **reverseproxy** |
| |`host`| The upstream host. |
| |`port`| The upstream port. |
| |`secret_token_path`| K8s Service Account Secret value. |
| |`k8s_dashboard_URL`| Kubernetes Dashboard. |
| **keycloak** |
| |`image`| keycloak Proxy Docker image. |

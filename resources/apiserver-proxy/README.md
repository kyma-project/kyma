# API Server Proxy

## Overview

This API Server Proxy is a transparent proxy for the Kubernetes API based on [kube-rbac-proxy](https://github.com/brancz/kube-rbac-proxy). It is exposed for the external communication.

## Details

Kyma requires all APIs, including those provided by the Kubernetes API server, to be exposed in a consistent manner through Istio.

To expose an API through Istio, all of the Pods that run the service containers must contain an Envoy sidecar. You need an additional proxy, as you cannot inject an Envoy sidecar directly into the Kubernetes API server. As a workaround, deploy apiserver-proxy as a proxy for the Kubernetes API server. Istio injects an Envoy sidecar into the Pods that run apiserver-proxy.

Installing the Helm chart creates a virtual service, which exposes the API server under the `apiserver` subdomain in the configured domain.

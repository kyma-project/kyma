# Kubernetes Dashboard proxy

## Overview

Kubernetes Dashboard proxy is a transparent proxy for the Kubernetes Dashboard allowing the Kyma clients to control the access to the Kubernetes Dashboard. Even though by default you don't need any verification to access the the Dashboard, the proxy uses Dex identity provider to ensure the Dashboard users are authenticated and authorized.

## Concept

The reverse proxy is a very thin layer Go application used by the Kubernetes Dashboard. It is used as an upstream for the [Keycloak Proxy](https://github.com/keycloak/keycloak-gatekeeper) which acts as a gateway for the Kubernetes Dashboard proxy.

The Keycloak proxy forwards the requests to Dex identity service, and asks you for authentication credentials. Based on the authentication credentials correctness, Dex checks if you are authorized to access the Kubernetes Dashboard.

The 1.8.1 version of Kubernetes Dashboard requires a second layer of authorization to get the full access to the Dashboard. You need to provide the HTTP authorization header with a Kubernetes service account secret value as a bearer token. The service account should be the same service account used to run the Kubernetes Dashboard proxy Pod, and must be bound to a Kubernetes cluster role allowing the access to the Dashboard resources. The application retrieves the Kubernetes services account secret, injects it as the value for the Authorization header and then forwards the HTTP request to the Kubernetes Dashboard.

## Docker images

Currently, the Kubernetes Dashboard proxy makes the following Docker image available to the Kyma Core Helm chart:

- k8s-dashboard-proxy

## Development

The main source code file for the Kubernetes Dashboard proxy resides under `cmd/reverseproxy`. It uses a Makefile to build the component and to create and push a Docker image. The following table explains the various `make` targets.

|Command| Description|
|-----------|------------|
|`make`|This is the default target for building the Docker image. It compiles, creates, and appropriately tags the Docker image.|
|`make compile`|It compiles the binary in the `bin` directory.|
|`make push`|Pushes the Docker image to the registry specified in the **REGISTRY** variable of the Makefile.|
|`make docker`|Creates the Docker image.|
|`make tag`|Tags the Docker image.|
|`make vet`|Runs `go vet` on all sources including **vendor** but excluding the `generated` directory.|

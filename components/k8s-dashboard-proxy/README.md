# Kubernetes Dashboard Proxy

## Overview

The `Kubernetes Dashboard Proxy` is a transparent proxy for the `Kubernetes Dashboard` to enable Kyma clients to control the access to the `Kubernetes Dashboard`. `Kubernetes Dashboard` by default is accessible with no authentication/authorization required. `Kubernetes Dashboard Proxy` leverages the Kyma Identity Provider (dex) to authenticate and authorize the `Kubernetes Dashboard` users.

## Concept

The idea is to introduce a reverse proxy in front of the Kubernetes Dashboard. The reverse proxy forwards the requests to the Authorization Server (`dex`) as it's the identity provider, asking the user for the authentication credentials. [Keycloak Proxy] (https://github.com/keycloak/keycloak-gatekeeper) is used here to be the gateway for the `Kubernetes Dashboard Proxy`.

Based on the authentication credentials correctness, `dex` checks if the user is authorized to access the Kubernetes Dashboard or not.
The currently used version of the `Kubernetes Dashboard` (*v1.8.1*), requires a second layer of authorization to get a full access to the dashboard which is providing the HTTP Authorization Header with a K8s Service Account Secret value as a Bearer Token. This Service Account should be the same Service Account who started the Kubernetes Dashboard Proxy pod and must be bound to a K8s Cluster Role allows the access to the     `Kubernetes Dashboard` resources.

A very thin layer Go application here is used as a reverse proxy in front of the `Kubernetes Dashboard` and is used as an upstream for the Keycloak Proxy. This thin layer app obtains the K8s Service Account Secret, injects it as a value for the Authorization header and then, forwards the HTTP request to the Kubernetes Dashboard.

## Docker Images

Currently, `Kubernetes Dashboard Proxy` makes the following Docker image available to the `kyma core` Helm chart:

- k8s-dashboard-proxy

## Development

The main source code file of `Kubernetes Dashboard Proxy` resides under `cmd/reverseproxy`. It has a Makefile to build the component as well as to create and push a Docker image. The following table explains the various make targets.


|Command| Description|
|-----------|------------|
|`make`|This is the default target for building the Docker image. It compiles, creates, and appropriately tags the Docker image.|
|`make compile`|It compiles the binary in the `bin` directory.|
|`make push`|Pushes the Docker image to the registry specified in the `REGISTRY` variable of the Makefile.|
|`make docker`|Creates the Docker image.|
|`make tag`|Tags the Docker image.|
|`make vet`|Runs `go vet` on all sources including `vendor` but excluding the `generated` directory.|

# etcd-tls-setup

## Overview

This image provides tools and scripts to generate TLS certificates for communication between the Service Catalog and etcd.

The etcd-tls-setup image has the following binaries installed:

* curl
* kubectl
* cfssl
* cfssljson

For more details, see this [Dockerfile](Dockerfile).

## Prerequisites

To set up the project, download these tools:

* [Go](https://golang.org/dl/) 1.11.4
* [Dep](https://github.com/golang/dep) v0.5.0
* [Docker](https://www.docker.com/)

These Go and Dep versions are compliant with the `buildpack` used by Prow. For more details read [this](https://github.com/kyma-project/test-infra/blob/master/prow/images/buildpack-golang/README.md) document.

## Usage

To build and run the etcd-tls-setup locally, run this command:

```bash
docker build -t etcd-tls-setup:latest . && docker run -it etcd-tls-setup:latest
```

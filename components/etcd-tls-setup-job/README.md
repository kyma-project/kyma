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

To set up the project, use these tools:
* Version 1.11.4 of [Go](https://golang.org/dl/)
* Version v0.5.0 of [Dep](https://github.com/golang/dep)
* The latest version of [Docker](https://www.docker.com/)

These versions are compliant with the `buildpack` used on Prow. For more information read [this](https://github.com/kyma-project/test-infra/blob/master/prow/images/buildpack-golang/README.md) document.

## Usage

To build and run the etcd-tls-setup locally, run this command:

```bash
docker build -t etcd-tls-setup:latest . && docker run -it etcd-tls-setup:latest
```
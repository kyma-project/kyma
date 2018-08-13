# etcd-tls-setup

## Overview

This image provides tools and scripts to generate TLS certificates for communication between the Service Catalog and etcd.

The etcd-tls-setup image has the following binaries installed:

* curl
* kubectl
* cfssl
* cfssljson

For more details, see this [Dockerfile](Dockerfile).

## Usage

To build and run the etcd-tls-setup locally, run this command:

```bash
docker build -t etcd-tls-setup:latest . && docker run -it etcd-tls-setup:latest
```
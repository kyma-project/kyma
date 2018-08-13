# etcd-tls

## Overview

This image provide tools and scripts to generate TLS certificates for communication between service catalog and etcd.

The etcd-tls image has following binaries installed:

* curl
* kubectl
* cfssl
* cfssljson

For more details see the [Dockerfile](Dockerfile)

## Usage

To build and run etcd-tls locally, call:

```bash
docker build -t etcd-tls:latest . && docker run -it etcd-tls:latest
```
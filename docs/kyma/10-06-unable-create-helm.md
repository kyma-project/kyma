---
title: Installation stuck at ContainerCreating
type: Troubleshooting
---

Starting with Kyma release 0.9.0 the communication with Helm and Tiller is [secured with TLS](/components/security/#details-tls-in-tiller).

If you try to install Kyma using your own image and the installation freezes at the `ContainerCreating` step, it means that the Kyma Installer cannot start because a required set of client-server certificates is not found in the system.

The `my-kyma.yaml` file contains two `image` fields. One of them defines the Tiller TLS certificates image and cannot be edited. Edit the field that defines the URL of the Kyma Installer image.

```
# This field defines the Tiller TLS certificates image URL. Do not edit.
image: eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20200121-4f3202bd

# This field defines the Kyma Installer image URL. Edit this field.
image: eu.gcr.io/kyma-project/develop/installer:0fdc80dd
```

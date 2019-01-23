# Configurations Generator

## Overview

The Kubeconfig generator is a proprietary tool that generates a `kubeconfig` file which allows the user to access the Kyma cluster through the Command Line Interface (CLI), and to manage the connected cluster within the permission boundaries of the user.

Read [this](../../../../docs/security/docs/03-01-kubecofig-generator.md) document to learn how to use the generator.

## The kubeconfig file

This is the format of the generated `kubecofig` file:

```
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: {CA_DATA}
    server: https://apiserver.{CLUSTER_DOMAIN}
  name: {CLUSTER_NAME_AND_DOMAIN}
contexts:
- context:
    cluster: {CLUSTER_NAME_AND_DOMAIN}
    user: OIDCUser
  name: {CLUSTER_NAME_AND_DOMAIN}
current-context: {CLUSTER_NAME_AND_DOMAIN}
kind: Config
preferences: {}
users:
- name: OIDCUser
  user:
    token: {TOKEN}
```

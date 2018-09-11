```
  ____             __ _                       _   _
 / ___|___  _ __  / _(_) __ _ _   _ _ __ __ _| |_(_) ___  _ __  ___
| |   / _ \| '_ \| |_| |/ _` | | | | '__/ _` | __| |/ _ \| '_ \/ __|
| |__| (_) | | | |  _| | (_| | |_| | | | (_| | |_| | (_) | | | \__ \
 \____\___/|_| |_|_| |_|\__, |\__,_|_|  \__,_|\__|_|\___/|_| |_|___/
                        |___/
  ____                           _
 / ___| ___ _ __   ___ _ __ __ _| |_ ___  _ __
| |  _ / _ \ '_ \ / _ \ '__/ _` | __/ _ \| '__|
| |_| |  __/ | | |  __/ | | (_| | || (_) | |
 \____|\___|_| |_|\___|_|  \__,_|\__\___/|_|

```

## Overview

This chart installs the Kubeconfig generator, a proprietary tool that generates a `Kubeconfig` file which allows the user to access to the Kyma cluster through the Command Line Interface (CLI), and to manipulate the connected cluster within the permission boundaries of the user.

Read [this](../../../../docs/authorization-and-authentication/docs/004-details-kubecofig-generator.md) document to learn how to use the generator.

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

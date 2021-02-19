# cert-manager

## Overview

`cert-manager` is a version of [Jetstack's cert-manager](https://cert-manager.io) customized to Kyma's needs.

Quoting from the [official README](https://github.com/jetstack/cert-manager):

>cert-manager is a Kubernetes addon to automate the management and issuance of
>TLS certificates from various issuing sources.
>
>It will ensure certificates are valid and up to date periodically, and attempt
>to renew certificates at an appropriate time before expiry.

## Changes introduced

The `crds.yaml` file has been split into multiple files. Each of these contains exactly one CRD with its name indicating the resource type. They are available in the `installation/resources/crds/cert-manager/` directory.

## Installation

To install `cert-manager`, use the Kyma Installer and follow these steps:

1. Manually apply CRDs from the project root directory:

   ```bash
   $ kubectl apply -f ./installation/resources/crds/cert-manager
   ```

2. Manually apply the Namespace resource from the project root directory:

   ```bash
   $ kubectl apply -f ./installation/resources/namespaces/cert-manager
   ```

3. Update the appropriate Installer CR template in `installation/resources/` by adding the following lines to the `spec.components` section:

    ```yaml
    - name: "cert-manager"
      namespace: "cert-manager"
    ```

## Usage

Before you configure `cert-manager` to issue certificates, you must first create Issuer resources. To learn more, read about [Issuers configuration](https://cert-manager.io/docs/configuration/).

One of the common use cases for `cert-manager` is securing ingress resources. Read the [Securing Ingresses Resources documentation](https://cert-manager.io/docs/usage/ingress/) for more details.

For more information, check the [official `cert-manager` documentation](https://cert-manager.io/docs/).

## Troubleshooting

One potential issue may occur while creating Issuer or ClusterIssuer resources. It might result in a failed call to the `webhook.cert-manager.io` webhook. A workaround for this is to delete these two `cert-manager`'s webhooks:

```bash
$ kubectl delete mutatingwebhookconfiguration.admissionregistration.k8s.io cert-manager-webhook
$ kubectl delete validatingwebhookconfigurations.admissionregistration.k8s.io cert-manager-webhook
```

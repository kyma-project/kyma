# cert-manager

## Overview

`cert-manager` is a Kyma's version of Jetstack's cert-manager.

Quote from the official README:
```
cert-manager is a Kubernetes addon to automate the management and issuance of
TLS certificates from various issuing sources.

It will ensure certificates are valid and up to date periodically, and attempt
to renew certificates at an appropriate time before expiry.
```

For more information check [links](#links) section.

## Changes introduced

Overall `crds.yaml` file has been split into multiple files. Each of these contains exactly one CRD with it's name indicating a resource type. They are available in `files/` directory.

## Prerequisites

- Kubernetes 1.11+

## Installing the Chart

In order to install cert-manager component using:

### Kyma installer
1. manually apply CRDs from `files/` directory
```bash
# Kubernetes 1.15+
$ kubectl apply -f ./files
```
2. update appropriate installer from `installation/resources/` and add the following section to `spec.components`:
```yaml
- name: "cert-manager"
  namespace: "cert-manager"
```

## Usage

[Issues configuration](https://cert-manager.io/docs/configuration/).

[Securing Ingresses documentation](https://cert-manager.io/docs/usage/ingress/).

For more information check [links](#links) section.

## Links

- [official site](https://cert-manager.io)
- [docs](https://cert-manager.io/docs/)
- [GitHub](https://github.com/jetstack/cert-manager)

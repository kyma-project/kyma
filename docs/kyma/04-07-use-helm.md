---
title: Use Helm
type: Installation
---

You can use Helm to manage Kubernetes resources in Kyma, for example to check the already installed Kyma charts or to install new charts that are not included in the Kyma Installer image.

## Helm v3

As of version 1.14, Kyma uses [Helm v3](https://helm.sh/) to install and maintain components. Unlike its predecessor, Helm v3 interacts directly with the Kubernetes API and thus no longer features an in-cluster server. With Tiller gone, managing Kubernetes resources using Helm v3 CLI requires no manual configuration.

## Helm v2

If you upgraded Kyma to v1.14, you can still fetch your pre-upgrade Helm v2 configuration and release data using Helm v2 CLI commands.
 
 >**CAUTION:** Do not use Helm v2 commands to modify existing Kyma components. Use them only to inspect pre-upgrade Kyma components or to modify custom components that have not been migrated to Helm v3.
 
Helm v2 relies on Tiller to govern charts and releases, so to use Helm v2 CLI, you must establish a secure connection with Tiller by saving the cluster's client certificate, key, and Certificate Authority (CA) to [Helm Home](https://v2.helm.sh/docs/glossary/#helm-home-helm-home) local directory.

>**TIP:** Read more about [TLS in Tiller](/components/security/#details-tls-in-tiller).

Run these commands at the end of the Kyma cluster installation to save the client certificate, key, and CA to [Helm Home](https://v2.helm.sh/docs/glossary/#helm-home-helm-home):

```bash
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem";
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem";
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem";
```

Additionally, you must add the `--tls` flag to every Helm command you run.

Helm v2 is a legacy mechanism. It will be removed in future releases.

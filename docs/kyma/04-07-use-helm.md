---
title: Use Helm
type: Installation
---

You can use Helm to manage Kubernetes resources in Kyma, for example to check the already installed Kyma charts or to install new charts that are not included in the Kyma Installer image.

To use it, you must establish a secure connection with Tiller by saving the cluster's client certificate, key, and Certificate Authority (CA) to [Helm Home](https://v2.helm.sh/docs/glossary/#helm-home-helm-home).

>**NOTE:** Read [this](/components/security/#details-tls-in-tiller) document to learn more about TLS in Tiller.

Run these commands at the end of the Kyma cluster installation to save the client certificate, key, and CA to [Helm Home](https://v2.helm.sh/docs/glossary/#helm-home-helm-home):

```bash
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem";
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem";
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem";
```

Additionally, you must add the `--tls` flag to every Helm command you run.

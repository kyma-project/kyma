---
title: '"Transport is closing" error'
type: Troubleshooting
---

Starting with Kyma release 0.9.0 the communication with Helm and Tiller is secured with TLS. If you get the `Transport is closing` error when you run a Helm command, such as `helm ls`, Helm denies you access because:

  - The cluster client certificate, key, and Certificate Authority (CA) are not found in [Helm Home](https://v2.helm.sh/docs/glossary/#helm-home-helm-home).
  - You don't use the `--tls` flag to engage a secure connection.

This problem is most common for cluster deployments where the user must add the required elements to Helm Home manually. When you install Kyma locally, this operation is performed automatically in the installation process.

Read [this](/components/security/#details-tls-in-tiller) document to learn more about security in communication with Helm and Tiller.

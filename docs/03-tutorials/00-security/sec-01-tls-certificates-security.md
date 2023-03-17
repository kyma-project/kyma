---
title: Set up a custom domain TLS certificate in Kyma
---

The TLS certificate is a vital security element. Follow this tutorial to set up your custom TLS certificate in Kyma.

>**NOTE:** This procedure can interrupt the communication between your cluster and the outside world for a limited period of time.

## Prerequisites

- Generate a new TLS certificate and key for custom domain deployments, base64-encoded.
- Generate a `kubeconfig` file for the Kubernetes cluster that hosts the Kyma instance.

## Steps
>**NOTE:** You can use this procedure to update your certificate after manual rotation.

1. Export your domain, new certificate, and key as the environment variables.

```bash
export DOMAIN={YOUR_DOMAIN}
export TLS_CERT={YOUR_NEW_CERTIFICATE}
export TLS_KEY={YOUR_NEW_KEY}
```

2. Deploy Kyma with your custom domain certificate. Run:

```bash
kyma deploy --domain $DOMAIN --tls-crt $TLS_CERT --tls-key $TLS_KEY
```

The process is complete when you see the `Kyma installed` message.

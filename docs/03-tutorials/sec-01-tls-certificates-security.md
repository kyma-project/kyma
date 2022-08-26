---
title: Set up custom domain certificates
---

This tutorial shows how to set up your custom TLS certificate, and how to allow mutual authentication in Kyma. 

## Set up your custom domain TLS certificate
The TLS certificate is a vital security element. In this section, you will learn how to set up your custom TLS certificate in Kyma.

>**NOTE:** This procedure can interrupt the communication between your cluster and the outside world for a limited period of time.

### Prerequisites

- New TLS certificate and key for custom domain deployments, base64-encoded
- `kubeconfig` file generated for the Kubernetes cluster that hosts the Kyma instance

### Steps

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



## Setup Kyma mTLS gateway with client certificates  

By default, [`kyma-gateway`](https://github.com/kyma-project/kyma/blob/main/resources/certificates/templates/certificate.yaml) doesn't support mutual authentication. In this section, you will learn how to setup [`kyma-mtls-gateway`](https://github.com/kyma-project/kyma/blob/main/resources/certificates/templates/mtls-certificate.yaml) to allow mutual authentication using client certificates issued by a trusted certificate authority (CA).

### Prerequisites

- Client root certificate issued by a trusted certificate authority (CA), base64-encoded, it could be a bundle certificate or a single certificate 
- `kubeconfig` file generated for the Kubernetes cluster that hosts the Kyma instance

### Steps

1. Export your client root CA as an environment variable.
   ```bash
   export CLIENT_ROOT_CA={YOUR_CLIENT_ROOT_CA}
   ```
2. Update the secret with the client root certificate. 
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: v1
   kind: Secret
   metadata:
   name: kyma-mtls-gateway-certs-cacert
   namespace: istio-system
   type: Opaque
   data:
   cacert: $CLIENT_ROOT_CA
   EOF
   ```
   The process is complete when you see the `secret/kyma-mtls-gateway-certs-cacert created` message.

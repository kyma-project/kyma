---
title: Rotate the Root certificate and the key issued by the Certificate Authority 
type: Tutorials
---

The Central Connector Service uses the Root certificate issued by the Certificate Authority (CA) to issue new certificates for Runtimes and by the Istio Ingress Gateway to validate their identity.

Two different components use the Root CA certificate. As a result, the certificate is stored in two separate Secrets:

  - The `connector-service-app-ca` Connector Service CA Secret responsible for signing certificate requests
  - The `kyma-gateway-certs-cacert` Istio Secret responsible for security in the Connector Service API

Keeping both Secrets up-to-date is vital for the security of your implementation as it guarantees that the Connector Service issues proper certificates and no unregistered Applications can access its API.

The Root CA certificate has a set expiration date and must be renewed periodically to prevent its expiration. You must also renew the Root CA certificate and key every time they are compromised.

This tutorial describes the procedure you must follow for these scenarios:

  - Rotating a soon-to-expire Root CA certificate
  - Rotating a compromised Root CA certificate
  - Rotating a compromised Root CA key

## Rotating a soon-to-expire CA certificate

To successfully rotate a soon-to-expire CA certificate, replace it with a new certificate in both the Connector Service CA Secret and the Istio Secret. Follow these steps to replace the old certificate in both Secrets:

1. Get the existing Root CA key. Fetch it from the `connector-service-app-ca` Secret and save it to the `ca.key` file.

   ```bash
   kubectl -n kyma-integration get secret connector-service-app-ca -o=jsonpath='{.data.ca\.key}' | base64 --decode > ca.key
   ```

1. Generate a new certificate for the key you obtained and save it to the `new-ca.crt` file.

   ```bash
   openssl req -new -key ca.key -x509 -sha256 -days {TTL_DAYS} -nodes -out new-ca.crt
   ```

   >**NOTE:** Use the `-days` flag to set the TTL (Time to live) of the newly generated certificate.

1. Encode the newly created certificate with base64.
  
   ```bash
   cat new-ca.crt | base64
   ```

1. Replace the old certificate in the Connector Service CA Secret. Edit the Secret and replace the `ca.crt` value with the new base64-encoded certificate.
  
   ```bash
   kubectl -n kyma-integration edit secret connector-service-app-ca
   ```

1. Get the existing Istio CA certificate. Fetch it from the `kyma-gateway-certs-cacert` Secret and save it to the `old-ca.crt` file.
  
   ```bash
   kubectl -n istio-system get secret kyma-gateway-certs-cacert -o=jsonpath='{.data.cacert}' | base64 --decode > old-ca.crt
   ```

1. Merge the old certificate and the newly generated certificate into a single `merged-ca.crt` file.
  
   ```bash
   cat old-ca.crt new-ca.crt > merged-ca.crt
   ```

1. Encode the newly created `merged-ca.crt` certificate file with base64.
  
   ```bash
   cat merged-ca.crt | base64
   ```

1. Replace the old certificate in the Istio Secret. Edit the Secret and replace the `cacert` value with the `merged-ca.crt` base64-encoded certificate.
  
   ```bash
   kubectl -n istio-system edit secret kyma-gateway-certs-cacert
   ```

1. Wait for all the client certificates to be renewed. 

    > **NOTE:** In the case of a Kyma Runtime connected to the central Connector Service, the system renews the certificates automatically using the information stored in the Secrets. Alternatively, you can renew the certificates in said Runtime yourself. To do that, create a CertificateRequest custom resource (CR) in the Runtime in which you want to renew the certificates.

1. After the client certificates are renewed, remove the `kyma-gateway-certs-cacert` Secret entry which contains the old certificate. First, encode the `new-ca.crt` file with base64.
  
   > **CAUTION:** Do not proceed with this step until all the client certificates are renewed!

   ```bash
   cat new-ca.crt | base64
   ```

1. Edit the Secret and replace the `cacert` value with the base64-encoded `new-ca.crt` certificate.
  
   ```bash
   kubectl -n istio-system edit secret kyma-gateway-certs-cacert
   ```

## Rotating a compromised Root CA certificate

1. Get the existing Root CA key. Fetch it from the `connector-service-app-ca` Secret and save it to a `ca.key` file.
  
   ```bash
   kubectl -n kyma-integration get secret connector-service-app-ca -o=jsonpath='{.data.ca\.key}' | base64 --decode > ca.key
   ```

2. Generate a new certificate for the key you obtained and save it to the `new-ca.crt` file.

   ```bash
   openssl req -new -key ca.key -x509 -sha256 -days {TTL_DAYS} -nodes -out new-ca.crt
   ```

   >**NOTE:** Use the `-days` flag to set the TTL (Time to live) of the newly generated certificate.

3. Encode the newly created certificate with base64.

   ```bash
   cat new-ca.crt | base64
   ```

4. Replace the old certificate in the Connector Service CA Secret. Edit the Secret and replace the `ca.crt` value with the new base64-encoded certificate.
  
   ```bash
   kubectl -n kyma-integration edit secret connector-service-app-ca
   ```

5. Replace the old certificate in the Istio Secret. Edit the Secret and replace the `ca.crt` value with the new base64-encoded certificate.
   
   ```bash
   kubectl -n istio-system edit secret kyma-gateway-certs-cacert
   ```

6. Generate new certificates in a Runtime. To do that, create a CertificateRequest custom resource (CR) in the Runtime in which you want to generate the certificates.

## Rotating a compromised Root CA key

1. Generate a new RSA-encoded root CA key and save it to the `new-ca.key` file.
   
   ```bash
   openssl genrsa -out new-ca.key 4096
   ```

2. Generate a new certificate using the key you generated and save it to the `new-ca.crt` file.

   ```bash
   openssl req -new -key new-ca.key -x509 -sha256 -days {EXPIRATION_DAYS} -nodes -out new-ca.crt
   ```

   >**NOTE:** Use the `-days` flag to set the TTL (Time to Live) of the newly generated certificate.

3. Encode the newly created certificate and key with base64:

   ```bash
   cat new-ca.crt | base64
   ```
  
   ```bash
   cat new-ca.key | base64
   ```

4. Replace the old certificate and key in the Connector Service CA Secret. Edit the Secret and replace the `ca.crt` and `ca.key` values with the new base64-encoded certificate and key respectively.
  
   ```bash
   kubectl -n kyma-integration edit secret connector-service-app-ca
   ```

5. Replace the old certificate in the Istio Secret. Edit the Secret and replace the `ca.crt` value with the new base64-encoded certificate.
  
   ```bash
   kubectl -n istio-system edit secret kyma-gateway-certs-cacert
   ```

6. Generate new certificates in a Runtime. To do that, create a CertificateRequest custom resource (CR) in the Runtime in which you want to generate the certificates.

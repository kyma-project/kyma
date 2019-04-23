---
title: Rotate the Root CA certificate and key
type: Tutorials
---

The Central Connector Service uses the Root CA certificate to issue new certificates for runtimes and by the Nginx Ingress Controller to validate their identity.

Two different components use the Root CA certificate. As a result, the certificate is stored in two separate Secrets:
  - The `connector-service-app-ca` Connector Service CA Secret responsible for signing certificate requests
  - The `nginx-auth-ca` Nginx Ingress Secret responsible for security in the Connector Service API

Keeping both Secrets up-to-date is vital for the security of your implementation as it guarantees that the Connector Service issues proper certificates and no unregistered applications can access its API.

The Root CA certificate has a set expiration date and must be renewed periodically to prevent its expiration. You must also renew the Root CA certificate and key every time they are compromised.

This tutorial describes the procedure you must follow for these scenarios:
  - Rotating a soon-to-expire Root CA certificate
  - Rotating a compromised Root CA certificate
  - Rotating a compromised Root CA key

## Rotating a soon-to-expire CA certificate

To successfully rotate a soon-to-expire CA certificate, replace it with a new certificate in both the Connector Service CA Secret and the Nginx Ingress Secret. Follow these steps to replace the old certificate in both Secrets:

1. Get the existing Root CA key. Fetch it from the `connector-service-app-ca` Secret and save it to a `ca.key` file. Run:
  ```
  kubectl -n kyma-integration get secret connector-service-app-ca -o=jsonpath='{.data.ca\.key}' | base64 --decode > ca.key
  ```

2. Generate a new certificate for the key you obtained and save it to a `new-ca.crt` file. Run:

  ```
  openssl req -new -key ca.key -x509 -sha256 -days {TTL_DAYS} -nodes -out new-ca.crt
  ```

>**NOTE:** Use the `-days` flag to set the TTL of the newly generated certificate.

3. Encode the newly created certificate with base64:
  ```
  cat new-ca.crt | base64
  ```

4. Replace the old certificate in the Connector Service CA Secret. Edit the Secret and replace the `ca.crt` value with the new base64-encoded certificate. Run:
  ```
  kubectl -n kyma-integration edit secret connector-service-app-ca
  ```

5. Get the existing Nginx Ingress Secret. Fetch it from the `nginx-auth-ca` Secret and save it to a `old-ca.crt` file. Run:
  ```
  kubectl -n kyma-integration get secret nginx-auth-ca -o=jsonpath='{.data.ca\.crt}' | base64 --decode > old-ca.crt
  ```

6. Merge the old Nginx certificate and the newly generated certificate into a single `nginx-ca.crt` file:
  ```
  cat old-ca.crt > nginx-ca.crt
  cat new-ca.crt >> nginx-ca.crt
  ```

7. Encode the newly created `nginx-ca.crt` certificate file with base64:
  ```
  cat nginx-ca.crt | base64
  ```

8. Replace the old certificate in the Nginx Ingress Secret. Edit the Secret and replace the `ca.crt` value with the `nginx-ca.crt` base64-encoded certificate. Run:
  ```
  kubectl -n kyma-integration edit secret nginx-auth-ca
  ```

9. Renew the certificates in a runtime. To do that, create a CertificateRequest CR in the runtime in which you want to renew the certificates. Alternatively, wait for the certificates to expire in a given runtime. The system renews the certificates automatically using the information stored in the Secrets you updated.


10. After the certificates are renewed in a runtime, remove the `nginx-auth-ca` Secret entry which contains the old certificate. First, encode the `new-ca.crt` file with base64:
  ```
  cat new-ca.crt | base64
  ```

11. Edit the Secret and replace the `ca.crt` value with the `new-ca.crt` base64-encoded certificate. Run:
  ```
  kubectl -n kyma-integration edit secret nginx-auth-ca
  ```

## Rotating a compromised Root CA certificate

1. Get the existing Root CA key. Fetch it from the `connector-service-app-ca` Secret and save it to a `ca.key` file. Run:
  ```
  kubectl -n kyma-integration get secret connector-service-app-ca -o=jsonpath='{.data.ca\.key}' | base64 --decode > ca.key
  ```

2. Generate a new certificate for the key you obtained and save it to a `new-ca.crt` file. Run:

  ```
  openssl req -new -key ca.key -x509 -sha256 -days {TTL_DAYS} -nodes -out new-ca.crt
  ```

>**NOTE:** Use the `-days` flag to set the TTL of the newly generated certificate.

3. Encode the newly created certificate with base64:
  ```
  cat new-ca.crt | base64
  ```

4. Replace the old certificate in the Connector Service CA Secret. Edit the Secret and replace `ca.crt` value with the new base64-encoded certificate. Run:
  ```
  kubectl -n kyma-integration edit secret connector-service-app-ca
  ```

5. Replace the old certificate in the Nginx Ingress Secret. Edit the Secret and replace the `ca.crt` value with the new base64-encoded certificate. Run:
  ```
  kubectl -n kyma-integration edit secret nginx-auth-ca
  ```

6. Renew the certificates in a runtime. To do that, create a CertificateRequest CR in the runtime in which you want to renew the certificates.

## Rotating a compromised root CA key

1. Generate a new, RSA-encoded root CA key and save to a `new-ca.key` file:
  ```
  openssl genrsa -out new-ca.key 2048
  ```

2. Generate a new certificate using the key you generated and save it to a `new-ca.crt` file. Run:

  ```
  openssl req -new -key ca.key -x509 -sha256 -days {EXPIRATION_DAYS} -nodes -out new-ca.crt
  ```

>**NOTE:** Use the `-days` flag to set the TTL of the newly generated certificate.

3. Encode the newly created certificate and key with base64:
  ```
  cat new-ca.key | base64
  ```
  ```
  cat new-ca.crt | base64
  ```

4. Replace the old certificate and key in the Connector Service CA Secret. Edit the Secret and replace the `ca.crt` and `ca.key` values with the new base64-encoded certificate and key respectively. Run:
  ```
  kubectl -n kyma-integration edit secret connector-service-app-ca
  ```

5. Replace the old certificate in the Nginx Ingress Secret. Edit the Secret and replace the `ca.crt` value with the new base64-encoded certificate. Run:
  ```
  kubectl -n kyma-integration edit secret nginx-auth-ca
  ```

6. Renew the certificates in a runtime. To do that, create a CertificateRequest CR in the runtime in which you want to renew the certificates.

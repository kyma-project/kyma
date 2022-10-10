---
title: Prepare self-signed client certificates
---

This tutorial shows how to create self-sign Root CA and use it to sign Client Certificate

## Prepare a Client Root CA

1. Export the following values as environment variables:

   ```bash
   export CLIENT_ROOT_CA_CN={ROOT_CA_COMMON_NAME}
   export CLIENT_ROOT_CA_ORG={ROOT_CA_ORGANIZATION}
   export CLIENT_ROOT_CA_KEY_FILE=${CLIENT_ROOT_CA_CN}.key
   export CLIENT_ROOT_CA_CRT_FILE=${CLIENT_ROOT_CA_CN}.crt
   ```

2. Generate a Client Root CA and a Client certificate:

   ```bash
   # Create a Root CA key and a certificate that's valid for one year - you can use it for validation
   openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=${CLIENT_ROOT_CA_ORG}/CN=${CLIENT_ROOT_CA_CN}' -keyout ${CLIENT_ROOT_CA_KEY_FILE} -out ${CLIENT_ROOT_CA_CRT_FILE}
   ```
   
## Prepare Client Certificate

1. Export the following values as environment variables:

   ```bash
   export CLIENT_CERT_CN={COMMON_NAME}
   export CLIENT_CERT_ORG={ORGANIZATION}
   export CLIENT_CERT_CRT_FILE=${CLIENT_CERT_CN}.crt
   export CLIENT_CERT_CSR_FILE=${CLIENT_CERT_CN}.csr
   export CLIENT_CERT_KEY_FILE=${CLIENT_CERT_CN}.key
   ```

2. Generate a Client certificate and sign it with the Root CA:

   ```bash
   # Create a new key and CSR for the client certificate
   openssl req -out ${CLIENT_CERT_CSR_FILE} -newkey rsa:2048 -nodes -keyout ${CLIENT_CERT_KEY_FILE} -subj "/CN=${CLIENT_CERT_CN}/O=${CLIENT_CERT_ORG}"
   # Sign the client certificate with the Client Root CA certificate
   openssl x509 -req -days 365 -CA ${CLIENT_ROOT_CA_CRT_FILE} -CAkey ${CLIENT_ROOT_CA_KEY_FILE} -set_serial 0 -in ${CLIENT_CERT_CSR_FILE} -out ${CLIENT_CERT_CRT_FILE}
   ```

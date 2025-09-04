# Prepare Self-Signed Root Certificate Authority and Client Certificates

This tutorial shows how to create a self-signed root certificate authority (CA) and how to use it to sign a client certificate.

> [!NOTE]
>  This solution is not recommended for production purposes. 


## Prepare a Client Root CA

1. Export the following values as environment variables:

   ```bash
   export CLIENT_ROOT_CA_CN={ROOT_CA_COMMON_NAME}
   export CLIENT_ROOT_CA_ORG={ROOT_CA_ORGANIZATION}
   export CLIENT_ROOT_CA_KEY_FILE=${CLIENT_ROOT_CA_CN}.key
   export CLIENT_ROOT_CA_CRT_FILE=${CLIENT_ROOT_CA_CN}.crt
   ```

2. Generate a client root CA and a client certificate:

   ```bash
   openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=${CLIENT_ROOT_CA_ORG}/CN=${CLIENT_ROOT_CA_CN}' -keyout ${CLIENT_ROOT_CA_KEY_FILE} -out ${CLIENT_ROOT_CA_CRT_FILE}
   ```
   
## Prepare a Client Certificate

1. Export the following values as environment variables:

   ```bash
   export CLIENT_CERT_CN={COMMON_NAME}
   export CLIENT_CERT_ORG={ORGANIZATION}
   export CLIENT_CERT_CRT_FILE=${CLIENT_CERT_CN}.crt
   export CLIENT_CERT_CSR_FILE=${CLIENT_CERT_CN}.csr
   export CLIENT_CERT_KEY_FILE=${CLIENT_CERT_CN}.key
   ```

2. Create a new key and CSR for the client certificate.
   
   ```bash
   openssl req -out ${CLIENT_CERT_CSR_FILE} -newkey rsa:2048 -nodes -keyout ${CLIENT_CERT_KEY_FILE} -subj "/CN=${CLIENT_CERT_CN}/O=${CLIENT_CERT_ORG}"
   ```

3. Sign the client certificate with the Client Root CA certificate.
   ```bash
   openssl x509 -req -days 365 -CA ${CLIENT_ROOT_CA_CRT_FILE} -CAkey ${CLIENT_ROOT_CA_KEY_FILE} -set_serial 0 -in ${CLIENT_CERT_CSR_FILE} -out ${CLIENT_CERT_CRT_FILE}
   ```

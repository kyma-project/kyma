# Application Connectivity Certs Setup Job

## Overview

Application Connectivity Certs Setup Job populates Kubernetes Secrets with the provided certificate and key pair or generates new one.


## Configuration

The Application Gateway has the following parameters:
**connectorCertificateSecret** - "Secret namespace/name where the Connector Service client certificate and key are kept.
**caCertificateSecret** - "Secret namespace/name where CA certificate is kept.
**caKey** - Specifies the base64-encoded private key for the Application Connector. If you don't provide it, a private key is generated automatically.
**caCertificate** - Specifies the base64-encoded certificate for the Application Connector. If you don't provide it, the certificate is generated automatically.
**generatedValidityTime** - Specifies how long the generated certificate is valid.


# Application Connectivity Certs Setup Job

## Overview

Application Connectivity Certs Setup Job populates Kubernetes Secrets with the provided certificate and key pair or generates new one.


## Configuration

The Application Gateway has the following parameters:
**connectorCertificateSecret** - Namespace and the name of the Secret which stores the Connector Service certificate and key. Requires the `Namespace/secret name` format.
**caCertificateSecret** - Namespace and the name of the Secret which stores the CA certificate. Requires the `Namespace/secret name` format.
**caKey** - Specifies the base64-encoded private key for the Application Connector. If you don't provide it, a private key is generated automatically.
**caCertificate** - Specifies the base64-encoded certificate for the Application Connector. If you don't provide it, the certificate is generated automatically.
**generatedValidityTime** - Specifies how long the generated certificate is valid.


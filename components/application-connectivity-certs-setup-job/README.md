# Application Connectivity Certs Setup Job

## Overview

Application Connectivity Certs Setup Job populates Kubernetes Secrets with the provided certificate and key pair or generates new one.


## Configuration

The Application Gateway has the following parameters:
**connectorCertificateSecret** - Namespace and the name of the Secret which stores the Connector Service certificate and key. Requires the `Namespace/secret name` format. <!-- TODO: to remove? -->
**caCertificateSecret** - Namespace and the name of the Secret which stores the CA certificate. Requires the `Namespace/secret name` format.
**caKey** - Specifies the base64-encoded private key for the Application Connector. If you don't provide it, a private key is generated automatically.
**caCertificate** - Specifies the base64-encoded certificate for the Application Connector. If you don't provide it, the certificate is generated automatically.
**caCertificateSecretToMigrate** - Namespace and the name of the Secret which stores the CA certificate to be renamed. Requires the `{NAMESPACE}/{SECRET_NAME}` format. 
**caCertificateSecretKeysToMigrate** - List of keys to be copied when migrating the old Secret specified in `caCertificateSecretToMigrate` to the new one specified in `caCertificateSecret`. Requires the JSON table format.
**connectorCertificateSecretToMigrate** - Namespace and the name of the Secret which stores the Connector Service certificate and key to be renamed. Requires the `{NAMESPACE}/{SECRET_NAME}` format.  <!-- TODO: to remove? -->
**connectorCertificateSecretKeysToMigrate** - List of keys to be copied when migrating the old Secret specified in `connectorCertificateSecretToMigrate` to the new one specified in `caCertificateSecret`. Requires the JSON table format. <!-- TODO: to remove? -->
**generatedValidityTime** - Specifies how long the generated certificate is valid for.

## Renaming secrets

To rename the Secret containing the CA cert, you must pass the following arguments:
- **caCertificateSecret** 
- **caCertificateSecretToMigrate** 
- **caCertificateSecretKeysToMigrate**

<!-- TODO: to remove? 
< To rename the Secret containing the CA cert and key, you must pass these arguments:
< - connectorCertificateSecret containing a new name
< - connectorCertificateSecretToMigrate containing an old name
< - connectorCertificateSecretKeysToMigrate containing a list of keys to copy from the old secret
-->

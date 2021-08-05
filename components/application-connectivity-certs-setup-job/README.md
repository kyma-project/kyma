# Application Connectivity Certs Setup Job

## Overview

Application Connectivity Certs Setup Job populates Kubernetes Secrets with the provided certificate and key pair or generates new one.


## Configuration

The Application Gateway has the following parameters:
**connectorCertificateSecret** - Namespace and the name of the Secret which stores the Connector Service certificate and key. Requires the `Namespace/secret name` format.
**caCertificateSecret** - Namespace and the name of the Secret which stores the CA certificate. Requires the `Namespace/secret name` format.
**caKey** - Specifies the base64-encoded private key for the Application Connector. If you don't provide it, a private key is generated automatically.
**caCertificate** - Specifies the base64-encoded certificate for the Application Connector. If you don't provide it, the certificate is generated automatically.
**caCertificateSecretToMigrate** - Namespace and the name of the Secret which stores the CA certificate to be renamed. Requires the `Namespace/secret name` format. 
**caCertificateSecretKeysToMigrate** - List of keys to be copied when migrating old secret specified in `caCertificateSecretToMigrate` to the new one specified in `caCertificateSecret`. Requires JSON table format.
**connectorCertificateSecretToMigrate** - Namespace and the name of the Secret which stores the Connector Service certificate and key to be renamed. Requires the `Namespace/secret name` format. 
**connectorCertificateSecretKeysToMigrate** - List of keys to be copied when migrating the old Secret specified in `connectorCertificateSecretToMigrate` to the new one specified in `caCertificateSecret`. Requires the JSON table format.
**generatedValidityTime** - Specifies how long the generated certificate is valid for.

## Renaming secrets

In order to rename secret containing CA cert the following arguments need to be passed:
- caCertificateSecret containing a new name
- caCertificateSecretToMigrate containing an old name
- caCertificateSecretKeysToMigrate containing a list of keys to copy from the old secret

In order to rename secret containing CA cert and key the following arguments need to be passed:
- connectorCertificateSecret containing a new name
- connectorCertificateSecretToMigrate containing an old name
- connectorCertificateSecretKeysToMigrate containing a list of keys to copy from the old secret

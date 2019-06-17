# Connector Service

## Overview

The Connector Service is responsible for generating and sending back client certificates based on Certificate Signing Request (CSR).

## Configuration

The Connector Service has the following parameters, that can be set through the chart:
- **appName** - This is the name of the application used by Kubernetes deployments and services. The default value is `connector-service`.
- **externalAPIPort** - This port exposes the Connector Service API to an external solution. The default port is `8081`.
- **internalAPIPort** - This port exposes the Connector Service within Kubernetes cluster. The default port is `8080`.
- **namespace** - Namespace where Connector Service is deployed. The default Namespace is `kyma-integration`.
- **tokenLength** - Length of registration tokens. The default value is `64`.
- **appTokenExpirationMinutes** - Time after which tokens for Applications expire and are no longer valid. The default value is `5` minutes.
- **runtimeTokenExpirationMinutes** - Time after which tokens for Runtimes expire and are no longer valid. The default value is `10` minutes.
- **caSecretName** - Namespace and the name of the Secret which stores the certificate and key used to sign client certificates. Requires the `Namespace/secret name` format. The default value is `kyma-integration/application-connector-certs-cacert`.
- **rootCACertificateSecretName** - Namespace and the name of the Secret which stores the root CA (Certificate Authority) Certificate. Requires the `Namespace/secret name` format.
- **requestLogging** - Flag for logging incoming requests. It is set to `False` by default.
- **connectorServiceHost** - Host under which this service is accessible. It is used for generating the URL. The default host is `cert-service.wormhole.cluster.kyma.cx`.
- **appRegistryHost** - Host under which the Application Registry is accessible. The default value is an empty string.
- **eventsHost** - Host under which the Event Service is accessible. The default value is an empty string.
- **appsInfoURL** - URL at which management information for applications is available. If not provided, it bases on `connectorServiceHost`.
- **runtimesInfoURL** - URL at which management information for runtimes is available. If not provided, it bases on `connectorServiceHost`.
- **group** - Group for which certificates are generated. If the chart does not provide a default value, you must specify it in the request header of the request sent to the token endpoint.
- **tenant** - Tenant for which certificates are generated. If the chart does not provide the default value, you must specify it in the header of the request sent to the token endpoint.
- **appCertificateValidityTime** - Time until which the certificates that the service issues for Applications are valid. The default value is 90 days.
- **runtimeCertificateValidityTime** - Time until which the certificates that the service issues for Runtimes are valid. The default value is 90 days.
- **central** - Determines whether the Connector Service works in the central mode.
- **revocationConfigMapName** - Name of the ConfigMap containing the revoked certificates list.

The Connector Service also uses the following environmental variables for CSR-related information config:
- **COUNTRY** (two-letter-long country code)
- **ORGANIZATION**
- **ORGANIZATIONALUNIT**
- **LOCALITY**
- **PROVINCE**

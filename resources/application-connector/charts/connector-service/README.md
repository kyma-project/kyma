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
- **appTokenExpirationMinutes** - Time after which tokens for applications expire and are no longer valid. The default value is `5` minutes.
- **runtimeTokenExpirationMinutes** - Time after which tokens for runtimes expire and are no longer valid. The default value is `10` minutes.
- **caSecretName** - Name of the secret which contains the root Certificate Authority (CA). The default value is `nginx-auth-ca`.
- **requestLogging** - Flag for logging incoming requests. It is set to `False` by default.
- **connectorServiceHost** - Host under which this service is accessible. It is used for generating the URL. The default host is `cert-service.wormhole.cluster.kyma.cx`.
- **gatewayHost** - Host at which gateway service is accessible The default value is `gateway.wormhole.cluster.kyma.cx`.
- **certificateProtectedHost** - Host secured with client certificate, used for certificate renewal. The default host is `gateway.wormhole.cluster.kyma.cx`.
- **appsInfoURL** - URL at which management information for applications is available. If not provided, it bases on `connectorServiceHost`.
- **runtimesInfoURL** - URL at which management information for runtimes is available. If not provided, it bases on `connectorServiceHost`.
- **certificateValidityTime** - Validity time of certificates issued by this service. The default value is 90 days.
- **central** - Determines whether the Connector Service works in the central mode.
- **revocationConfigMapName** - Name of the config map containing revoked certificates list.

The Connector Service also uses the following environmental variables for CSR-related information config:
- **COUNTRY** (two-letter-long country code)
- **ORGANIZATION**
- **ORGANIZATIONALUNIT**
- **LOCALITY**
- **PROVINCE**

# Connector Service

## Overview

This is the repository for the Kyma Connector Service.

## Prerequisites

The Connector Service requires Go 1.8 or higher.

## Usage

This section explains how to use the Connector Service.

### Start the Connector Service
To start the Connector Service, run this command:

```
./connectorservice
```

The Connector Service has the following parameters:
- **appName** - This is the name of the application used by k8s deployments and services. The default value is `connector-service`.
- **externalAPIPort** - This port exposes the Connector Service API to an external solution. The default port is `8081`.
- **internalAPIPort** - This port exposes the Connector Service within Kubernetes cluster. The default port is `8080`.
- **namespace** - Namespace where Connector Service is deployed. The default namespace is `kyma-integration`.
- **tokenLength** - Length of registration tokens. The default value is `64`.
- **tokenExpirationMinutes** - Time after which tokens expire and are no longer valid. The default value is `60` minutes.
- **domainName** - Domain name of the cluster, used for URL generating. Default domain name is `.wormhole.cluster.kyma.cx`.
- **certificateServiceHost** - Host at which this service is accessible, used for URL generating. Default host is `cert-service.wormhole.cluster.kyma.cx`.

Connector Service also uses following environmental variables for CSR - related information config:
- **COUNTRY** (two-letter-long country code)
- **ORGANIZATION**
- **ORGANIZATIONALUNIT**
- **LOCALITY**
- **PROVINCE**

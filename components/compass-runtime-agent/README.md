# Compass Runtime Agent

## Overview

This is the repository for the Compass Runtime Agent.

The main responsibilities of the Compass Runtime Agent are:
- Establishing a trusted connection between the Runtime and the Compass
- Renewing the trusted connection between the Runtime and the Compass
- Configuring the Runtime


## Prerequisites

The Compass Runtime Agent requires Go 1.8 or higher.

## Usage

To start the Compass Runtime Agent, run this command:

```
./compass-runtime-agent
```

## Parameters and environment variables

The Compass Runtime Agent has the following parameters:
- **controllerSyncPeriod** is the time period between resyncing existing resources. Provide it in seconds. The default value is `60`.
- **minimalConfigSyncTime** is the minimal time between synchronizing the configuration. Provide it in seconds. The default value is `300`.
- **integrationNamespace** is the Namespace in which the resources are created. The default Namespace is `kyma-integration`.
- **gatewayPort** is the Application Gateway port. The default port is `8080`.
- **insecureConfigurationFetch** specifies whether to fetch the configuration with disabled TLS verification. The default value is `false`.
- **uploadServiceUrl** is the URL of the upload service. By default, it is an empty string.

The Compass Runtime Agent uses these environment variables:
- **DIRECTOR_URL**
- **RUNTIME_ID**
- **TENANT**

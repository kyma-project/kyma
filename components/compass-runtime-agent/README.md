# Compass Runtime Agent

## Overview

This is the repository for the Kyma Compass Runtime Agent.

## Prerequisites

The Compass Runtime Agent requires Go 1.8 or higher.

## Usage

This section explains how to use the Compass Runtime Agent.

### Start the Compass Runtime Agent
To start the Connector Service, run this command:

```
./compass-runtime-agent
```

The Compass Runtime Agent has the following parameters:
- **controllerSyncPeriod** is the time period between resyncing exissting resources. 
- **minimalConfigSyncTime** is the minimal time between synchronizing the configuration.
- **integrationNamespace** is the namespaces in which the resources are created.
- **gatewayPort** is the Application Gateway port.
- **insecureConfigurationFetch** specifies whether to fetch the configuration with disabled TLS verification.
- **uploadServiceUrl** is the URL of the upload service.

The Compass Runtime Agent also uses the following environment variables:
- **DIRECTOR_URL**
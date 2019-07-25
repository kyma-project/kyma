# Compass

## Overview

Compass is a multi-tenant system which consists of components that provide a way to register, group, and manage your applications across multiple Kyma runtimes. Compass consists of the following sub-charts:

- `connector` 
- `director` 
- `gateway` 
- `healthchecker`
- `postgresql`

To learn more, read the [Compass documentation](https://github.com/kyma-incubator/compass/blob/master/README.md).
## Details

### Configuration
The following table lists the configurable parameters of the Compass chart and their default values.

| Parameter | Description | Values | Default |
| --- | --- | --- | --- |
| **database.useEmbedded** | Specifies whether `postgresql` chart should be installed. | true/false | `true` |

To learn how to use managed GCP database, see the [Configure Managed GCP PostgreSQL](./configure-managed-gcp-postgresql.md) document.

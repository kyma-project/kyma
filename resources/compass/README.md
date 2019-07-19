# Compass

## Overview

The Compass consists of the following sub-charts:

- `connector` 
- `director` 
- `gateway` 
- `healthchecker`
- `postgresql`

## Details

To learn more about the Compass, see the [Overview](https://github.com/kyma-incubator/compass#overview) document.

## Configuration

| Parameter | Description | Values | Default |
| --- | --- | --- | --- |
| `database.useEmbedded` | Specifies whether `postgresql` chart should be installed | true/false | `true` |

To learn how to use managed GCP database see [this document](https://github.com/kyma-incubator/compass/blob/master/docs/configure-managed-gcp-postgresql.md)
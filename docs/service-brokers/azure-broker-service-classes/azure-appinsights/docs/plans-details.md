---
title: Services and Plans
type: Details
---

## Service description

The `azure-appinsights` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `asp-dot-net-web` | For ASP.NET web applications. |
| `java-web` | For Java web applications. |
| `node-dot-js` | For Node.JS applications. |
| `general` | For general applications. |
| `app-center` | For Mobile applications. |

## Provision

Provisions a new Application Insights.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **location** | `string` | The Azure region in which to provision applicable resources. | Yes |  |
| **resourceGroup** | `string` | The (new or existing) resource group with which to associate new resources. | Yes |  |
| **appInsightsName** | `string` | The Application Insights component name. | No | A randomly generated UUID. |
| **tags** | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |

## Bind

Returns the instrumentation key.

### Binding Parameters

This binding operation does not support any parameters.

### Credentials

Binding returns the following connection details and shared credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| **instrumentationKey** | `string` | Instrumentation key. |

## Deprovision

Deletes the Application Insights.
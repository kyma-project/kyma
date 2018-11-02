---
title: Classes
type: UI OSBA Contracts
---

# Contract with Open Broken Service API

Catalog-ui directly uses the [ui-api-layer](https://github.com/kyma-project/kyma/tree/master/components/ui-api-layer) project to fetch the data. The ui-api-layer fetches the data from Service Brokers using the Service Catalog. The next section explains the mapping of [Service Object](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#catalog-management) from [OSBA](https://openservicebrokerapi.org/) to UI fields.

## Catalog page

| Number | OSBA field                        | Fallbacks  | Description                                                                                                                |
| ------ | --------------------------------- | ---------- | -------------------------------------------------------------------------------------------------------------------------- |
| (1)    | metadata.displayName              | name*, id* | If **metadata.displayName, name, id** fields are not present, the given Service Class does not appear on the landing page. |
| (2)    | metadata.providerDisplayName      | -          | If not provided, UI displays without this information.                                                                     |
| (3)    | description\*                     | -          | If not provided, UI displays without this information.                                                                     |
| (4)    | metadata.labels\*\*               | -          | If not provided, UI does not display any labels.                                                                           |
| (5)    | metadata.labels.local\*\* and/or metadata.labels.showcase\*\* | - | If not provided, choosing Basic Filter is not possible.                                                 |
| (6)    | tags                              | -          | If not provided, filtering by Tag is not possible.                                                                         |
| (7)    | metadata.labels.connected-app\*\* | -          | If not provided, choosing Connected Applications is not possible.                                                          |
| (8)    | metadata.providerDisplayName      | -          | If not provided, filtering by Provider is not possible.                                                                    |

\*Fields with an asterisk are required OSBA attributes.

\*\*`metadata.labels` is the custom object that is not defined in the [OSBA metadata convention](https://github.com/openservicebrokerapi/servicebroker/blob/master/profile.md#service-metadata)

![alt text](./assets/screen-catalog-page.png 'Catalog')

## Catalog Details page

| Number | OSBA field                   | Fallbacks      | Description                                                       |
| ------ | ---------------------------- | -------------- | ----------------------------------------------------------------- |
| (1)    | metadata.displayName         | name*, id*     | -                                                                 |
| (2)    | metadata.providerDisplayName | -              | If not provided, UI displays without this information.            |
| (3)    | not related to OSBA          | -              | -                                                                 |
| (4)    | metadata.documentationUrl    | -              | If not provided, the link with documentation does not appear.     |
| (5)    | metadata.supportUrl          | -              | If not provided, the link with support does not appear.           |
| (6)    | tags                         | -              | If not provided, UI displays without tags.                        |
| (7)    | metadata.longDescription     | description\*  | If not provided, the `General Information` panel does not appear. |
| (8)    | not related to OSBA          | -              | -                                                                 |

\*Fields with an asterisk are required OSBA attributes.

![alt text](./assets/screen-catalog-details-page.png 'Catalog Details')

## Add to Environment

| Number | OSBA field                | Fallbacks            | Description |
| ------ | ------------------------- | -------------------- | ----------- |
| (1)    | plan.metadata.displayName | plan.name*, plan.id* |             |
| (2)    | not related to OSBA       | -                    |             |
| (3)    | not related to OSBA       | -                    |             |

\*Fields with an asterisk are required OSBA attributes.

![alt text](./assets/screen-add-to-environment.png 'Add to Environment')

### Plan schema

[Plan Object](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#schema-object) in the OSBA can have the **schemas** field. Schema is used to generate form to enable provisioning of the Service Class.

Example:

```
{
          "$schema": "http://json-schema.org/draft-04/schema#",
          "properties": {
            "imagePullPolicy": {
              "default": "IfNotPresent",
              "enum": [
                "Always",
                "IfNotPresent",
                "Never"
              ],
              "title": "Image pull policy",
              "type": "string"
            },
            "redisPassword": {
              "default": "",
              "format": "password",
              "description": "Redis password. Defaults to a random 10-character alphanumeric string.",
              "title": "Password (Defaults to a random 10-character alphanumeric string)",
              "type": "string"
            }
          },
          "type": "object"
        }
```

Form:

![alt text](./assets/screen-schema-form.png 'SchemaForm')

Best practices for designing schema object:

* If the field has limited possible values, use the **enum** field. It renders a field as dropdown, so it prevents user from making the mistakes.
* If the field is required for the Service Class, mark it as **required**. UI blocks provisioning if you do not fill in the required fields.
* Fill the **default** value for a field whenever possible, it makes the provisioning faster.
* If the field, such as the field for a password, must be starred, use the **format** key with the **password** value.

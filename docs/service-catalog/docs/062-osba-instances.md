---
title: Instances
type: UI OSBA Contracts
---

Instances UI directly uses the [UI API Layer](https://github.com/kyma-project/kyma/tree/master/components/ui-api-layer) project to fetch the data. The UI API Layer fetches the data from Service Brokers using the Service Catalog. The next section explains the mapping of [Service Object](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#catalog-management) from [OSBA](https://openservicebrokerapi.org/) to UI fields.

## Service Instances page

| Number | OSBA field                | Fallbacks            | Description                                                                  |
| ------ | ------------------------- | -------------------- | ---------------------------------------------------------------------------- |
| (1)    | not related to OSBA       | -                    | It is the name of the service instance, created during service provisioning. |
| (2)    | **metadata.displayName**      | **name*, ** **id* **           | If not provided, UI displays without this information.                       |
| (3)    | **plan.metadata.displayName** | **plan.name*, ** **plan.id* **| If not provided, UI displays without this information.                       |
| (4)    | not related to OSBA       | -                    |                                                                              |
| (5)    | not related to OSBA       | -                    |                                                                              |
|        |

\*Fields with an asterisk are required OSBA attributes.

![alt text](./assets/screen-instances.png 'Service Instances')

## Service Instance Details page

| Number | OSBA field                | Fallbacks            | Description                                           |
| ------ | ------------------------- | -------------------- | ----------------------------------------------------- |
| (1)    | **metadata.displayName**      | **name* **, **id* **           | -                                                     |
| (2)    | **plan.metadata.displayName** | **plan.name*, ** **plan.id* ** | -                                                     |
| (3)    | not related to OSBA       | -                    | If not provided, UI displays without this information |
| (4)    | **metadata.documentationUrl** | -                    | If not provided, UI displays without this information |
| (5)    | **metadata.supportUrl**       | -                    | If not provided, UI displays without this information |
| (6)    | **description\* **             | -                    | If not provided, UI displays without this information |
| (7)    | not related to OSBA       | -                    | -                                                     |

\*Fields with an asterisk are required OSBA attributes.

![alt text](./assets/screen-instances-details.png 'Service Instance Details')

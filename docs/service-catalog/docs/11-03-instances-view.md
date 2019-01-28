---
title: Instances View
type: UI Contracts
---

This document describes the mapping of [OSBA service objects](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#service-objects), [plan objects](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#plan-object), and [conventions](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/profile.md#service-metadata) in the Kyma Console Instances view.

## Service Instances page

These are the OSBA fields used in the main Instances page:

| Number | OSBA field                | Fallbacks            | Description                                                                  |
| ------ | ------------------------- | -------------------- | ---------------------------------------------------------------------------- |
| (1)    | not related to OSBA       | -                    | It is the name of the Service Instance, created during service provisioning. |
| (2)    | **metadata.displayName**      | **name***, **id***           | If not provided, UI does not display this information.                       |
| (3)    | **plan.metadata.displayName** | **plan.name***, **plan.id***| If not provided, UI does not display this information.                       |
| (4)    | not related to OSBA       | -                    |                                                                              |
| (5)    | not related to OSBA       | -                    |                                                                              |
|        |

\*Fields with an asterisk are required OSBA attributes.

![alt text](./assets/instances.png 'Service Instances')

## Service Instance Details page

These are the OSBA fields used in the detailed Service Instance view:

| Number | OSBA field                | Fallbacks            | Description                                           |
| ------ | ------------------------- | -------------------- | ----------------------------------------------------- |
| (1)    | **metadata.displayName**      | **name***, **id***           | -                                                     |
| (2)    | **plan.metadata.displayName** | **plan.name***, **plan.id*** | -                                                     |
| (3)    | **metadata.documentationUrl** | -                    | If not provided, UI does not display this information |
| (4)    | **metadata.supportUrl**       | -                    | If not provided, UI does not display this information |
| (5)    | **description\***             | -                    | If not provided, UI does not display this information |

\*Fields with an asterisk are required OSBA attributes.

![alt text](./assets/instances-details.png 'Service Instance Details')

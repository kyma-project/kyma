---
title: Brokers view
type: UI Contracts
---

This document describes the mapping of [OSBA service objects](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#service-objects) in the Kyma Console Brokers view.

## Service Brokers page

These are the OSBA fields used in the main Brokers page:

| Number | OSBA field                | Fallbacks            | Description                                                                  |
| ------ | ------------------------- | -------------------- | ---------------------------------------------------------------------------- |
| (1)    | not related to OSBA       | -                    | Name of the Service Broker.                                                  |
| (2)    | not related to OSBA       | -                    | Age of the Service Broker.                                                   |
| (3)    | **spec.URL**              | -                    | URL of the Service Broker.                                     |
| (4)    | not related to OSBA       | -                    | Status of the Service Broker.                                                |
|        |

![alt text](assets/service-brokers.png 'Service Brokers')

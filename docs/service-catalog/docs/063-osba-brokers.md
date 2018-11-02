---
title: Brokers
type: UI OSBA Contracts
---

# Contract with the Open Service Broker API

Brokers UI directly uses the [UI API Layer](https://github.com/kyma-project/kyma/tree/master/components/ui-api-layer) project to fetch the data. The UI API Layer fetches the data from the Service Brokers using the Service Catalog. The next section explains the [Service Object](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#catalog-management) mapping from the [Open Service Broker API](https://openservicebrokerapi.org/) (OSBA) to UI fields.

## Service Brokers page

| Number | OSBA field                | Fallbacks            | Description                                                                  |
| ------ | ------------------------- | -------------------- | ---------------------------------------------------------------------------- |
| (1)    | not related to OSBA       | -                    | Name of the Service Broker.                                                  |
| (2)    | not related to OSBA       | -                    | Age of the Service Broker.                                                   |
| (3)    | **spec.URL**              | -                    | URL of the Service Broker.                                     |
| (4)    | not related to OSBA       | -                    | Status of the Service Broker.                                                |
|        |

![alt text](assets/service-brokers.png 'Service Brokers')

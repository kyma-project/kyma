---
title: Overview
type: UI Contracts
---

UI Contracts are contracts between the Service Catalog views in the Kyma Console UI and the [Open Service Broker API](https://www.openservicebrokerapi.org/) (OSBA) specification.

There are three types of OSBA fields:
* Mandatory fields which are crucial to define
* Optional fields which you can but do not have to define
* Conventions which are proposed fields that can be passed in the **metadata** object

The Service Catalog is OSBA-compliant, which means that you can register a Service Class that has only the mandatory fields.
However, it is recommended to provide more detailed Service Class definitions for better user experience.

In the Kyma Console UI, there are two types of views:
- Catalog view
- Instances view

Read the [Catalog view](#ui-contracts-catalog-view) and [Instances view](#ui-contracts-instances-view) documents to:
- Understand the contract mapping between the Kyma Console UI and the OSBA
- Learn which fields are primary to define, to provide the best user experience
- See which fields are used as fallbacks if you do not provide the primary ones

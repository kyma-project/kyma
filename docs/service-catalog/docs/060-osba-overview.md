---
title: Overview
type: UI Contracts
---

UI Contracts are contracts between the Service Catalog views in the Kyma Console UI and the [Open Service Broker API](https://www.openservicebrokerapi.org/) (OSBA) specification.

There are three types of OSBA fields:
* mandatory fields, which are crucial to display the page content properly
* optional fields, which you can but do not have to define
* conventions, which are proposed fields that can be changed with your own fallbacks

The Service Catalog in Kyma is OSBA-compliant, which means that you can register a Service Class that has only the mandatory fields defined. Such Service Class displays in the Catalog view, however, it does not display in a pleasant and satisfactory way. Follow the UI Contracts and write your definitions in the recommended way to provide the best user experience.

The are three views in the Kyma Console UI:
- Catalog view
- Instances view
- Brokers view

Read the **Catalog view**, **Instances view**, and **Brokers view** documents to learn more about the OSBA fields used in those particular Kyma Console UI views.

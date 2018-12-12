---
title: Application Broker
type: Overview
---

The Application Broker (AB) provides Applications in the Service Catalog. An Application represents an external solution connected to the Kyma instance. The AB enables the integration of independent Applications within Kyma. It also allows you to extend the functionality of existing systems.

The AB observes all the Application custom resources and exposes their APIs and Events as ServiceClasses to the Service Catalog. When the list of remote the ServiceClasses of an Application is available in the Service Catalog, you can create an ApplicationMapping, provision those ServiceClasses, and enable them for Kyma services.

The AB implements the [Open Service Broker API](https://www.openservicebrokerapi.org/). For more details about the Service Brokers, see the Service Brokers **Overview** documentation.

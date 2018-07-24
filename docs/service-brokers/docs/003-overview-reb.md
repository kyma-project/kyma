---
title: Remote Environment Broker
type: Overview
---

The Remote Environment Broker (REB) provides remote environments in the Service Catalog. A remote environment represents the environment connected to the Kyma instance. The Remote Environment Broker enables the integration of independent remote environments within Kyma. It also allows you to extend the functionality of existing systems.

The REB observes all the remote environment custom resources and exposes their APIs and/or Events as ServiceClasses to the Service Catalog. When the list of remote environments' ServiceClasses is available in the Service Catalog, you can create an EnvironmentMapping, provision those ServiceClasses, and enable them for Kyma services.

The REB implements the Service Broker API. For more details about the Service Brokers, see the Service Brokers **Overview** documentation.

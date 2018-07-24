---
title: Consuming applications through the Service Catalog
type: Details
---

To consume the external solutions, referred to as Remote Environments, register them in Kyma. As a result of registering the external solutions, ClusterServiceClasses are created in the Service Catalog.

### How an external solution is represented in the Service Catalog
This document presents the example referring to the Order API ClusterServiceClass. This class is registered in Kyma with a `targetUrl` pointing to `https://www.orders.com/v1/orders`. The response `id` during the registration is `01a702b8-e302-4e62-b678-8d361b627e49`.

As a result, the Remote Environment Broker, which provides ServiceClasses to the Service Catalog, contains the class with the following `id`:
```
re-{remote-environment-name}-gateway-{service-id}
```
The `{service-id}` is an identifier returned in the process of registration. The `{remote-environment-name}` is the name of the Remote Environment created in Kyma. It represents an instance of the external solution that owns the registered service. Such an `id` in the Service Broker is referred to as a `name` of the ClusterServiceClass in the Service Catalog.
Example `name`:
```
re-ec-default-gateway-01a702b8-e302-4e62-b678-8d361b627e49
```

### Service consumption

After provisioning the Order API in the environment using the Service Catalog, you can bind it to your application and consume it by calling the `url` provided during the binding operation.

The following example shows the Gateway `url` provided for your applications:
```
re-ec-default-gateway-01a702b8-e302-4e62-b678-8d361b627e49.kyma-integration/orders
```
The Gateway proxies all your requests to `https://www.orders.com/v1/orders`, in the case of the Order API example. You do not have to obtain the OAuth token in your application to access the API because the Gateway does it for you.

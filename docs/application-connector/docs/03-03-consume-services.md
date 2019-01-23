---
title: Consume applications through the Service Catalog
type: Details
---

To consume an external solution connected to Kyma, you must register it as an Application (App). As a result of registering the external solution, ClusterServiceClasses are created in the Service Catalog.

## External solution's services in the Service Catalog

The Example API is registered in Kyma with the `targetUrl` pointing to `https://www.orders.com/v1/orders`. The ID assigned to the API in the registration process is `01a702b8-e302-4e62-b678-8d361b627e49`.

The Application Broker, which provides ServiceClasses to the Service Catalog, follows this naming convention for its objects:
```
app-{application-name}-{service-id}
```
The `{service-id}` is the service identifier assigned in the process of registration. The `{application}` is the name of the App created in Kyma. It represents an instance of the connected external solution that owns the registered service. Such identifier used by the Application Broker is referred to as the `name` of a ClusterServiceClass in the Service Catalog.

This an example of a full ClusterServiceClass `name`:
```
re-ec-default-01a702b8-e302-4e62-b678-8d361b627e49
```

## Service consumption

After you provision the Example API in the Namespace of your choice using the Service Catalog, you can bind it to your application and consume it by calling the URL you get as a result of a successful binding.

This is a sample URL for the Example API:
```
re-ec-default-01a702b8-e302-4e62-b678-8d361b627e49.kyma-integration/orders
```

When you call this URL, the Application Proxy passes all requests to the `https://www.orders.com/v1/orders` address, which is the `targetUrl` registered for the Example API. You do not have to get an OAuth token and manually include it in the call as the Application Proxy does it for you automatically.

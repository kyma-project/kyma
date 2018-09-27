---
title: Managing services in Metadata API
type: Getting Started
---
The process of connecting your external system to Kyma consists of two steps:
- Obtaining the certificate
- Registering services

In this section we will focus on the second step, the information on how to
obtain proper certificate are available in previous section of this Getting
Started doc.

## Prerequisites

The only prerequisite to follow this guide is to have a valid certificate signed
by Kyma's Certificate Authority

Gateway Service and Event Services are exposed via `core-nginx-ingress-controller`
and thus you need to use it's port in requests, you can obtain it via:

```
export NODE_PORT=`kubectl -n kyma-system get svc core-nginx-ingress-controller -o 'jsonpath={.spec.ports[?(@.port==443)].nodePort}'`
```

After executing this command it will be available in NODE_PORT environment variable.

## Registering a service

At this point your Remote Environment has no registered services, you can check
that by making following call to Metadata Service:
```
curl https://gateway.kyma.local:$NODE_PORT/{REMOTE_ENVIRONMENT_NAME}/v1/metadata/services --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
```

In order to register a service you will need to prepare a body for the request
describing your service, you can check details in [Metadata Service API reference](TODO)
or use following example:
```
{
  "provider": "example-provider",
  "name": "example-name",
  "description": "this is long description of your service",
  "shortDescription": "very brief description",
  "labels": {
    "example": "true",
  },
  "api": {
    "targetUrl": "https://httpbin.org/",
    "spec": {}
  },
  "events": {
    "spec": {TODO}
  },
  "documentation": {
    "displayName": "Documentation",
    "description": "Description",
    "type": "some type",
    "tags": ["tag1", "tag2"],
    "docs": [
        {
        "title": "Documentation title...",
        "type": "type",
        "source": "source"
        }
    ]
  }
}
```

Then, just make the following call:
```
TODO
```

And the response should look like:
```
TODO
```

## Updating a service

In order to make an update to the existing service prepare new body with desired
values and make a following request:
```
TODO
```

## Deleting a service

To delete a service simply make a DELETE call to the following endpoint:
```
TODO
```

Having your service registered in Metadata Service allows you to send events
and consume it's API from within Kyma. The guides on how to do so are available
in latter sections of our Getting Started guide.
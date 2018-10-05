---
title: Register and manage services
type: Getting Started
---

This guide shows you how to register and manage the services of your external solution in Kyma.

## Prerequisites

Valid certificate signed by the Kyma Certificate Authority.

## Register a service

To register a service, follow this template to prepare the request body:
```
{
  "provider": "example-provider",
  "name": "example-name",
  "description": "This is the long description of your service",
  "shortDescription": "Short description",
  "labels": {
    "example": "true"
  },
  "api": {
    "targetUrl": "https://httpbin.org/",
    "spec": {}
  },
  "events": {
    "spec": {
      "asyncapi": "1.0.0",
      "info": {
        "title": "PetStore Events",
        "version": "1.0.0",
        "description": "Description of all the events"
      },
      "baseTopic": "stage.com.some.company.system",
      "topics": {
        "petCreated.v1": {
          "subscribe": {
            "summary": "Event containing information about new pet added to the Pet Store.",
            "payload": {
              "type": "object",
              "properties": {
                "pet": {
                  "type": "object",
                  "required": [
                    "id",
                    "name"
                  ],
                  "example": {
                    "id": "4caad296-e0c5-491e-98ac-0ed118f9474e",
                    "category": "mammal",
                    "name": "doggie"
                  },
                  "properties": {
                    "id": {
                      "title": "Id",
                      "description": "Resource identifier",
                      "type": "string"
                    },
                    "name": {
                      "title": "Name",
                      "description": "Pet name",
                      "type": "string"
                    },
                    "category": {
                      "title": "Category",
                      "description": "Animal category",
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
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

Include the request body you prepared in this call:
```
curl -X POST -d '{YOUR_REQUEST_BODY}' https://gateway.kyma.local:$NODE_PORT/ec-default/v1/metadata/services --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
```

A successful response returns the ID of the registered service:
```
{"id":"{YOUR_SERVICE_ID}"}
```

To check the details of a registered service, send this request:
```
curl https://gateway.kyma.local:$NODE_PORT/ec-default/v1/metadata/services/{YOUR_SERVICE_ID} --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
```

## Update a service

To update an existing service, prepare a new request body using the template from the following step and send it in this request:
```
curl -X PUT -d '{YOUR_NEW_REQUEST_BODY}' https://gateway.kyma.local:$NODE_PORT/ec-default/v1/metadata/services/{YOUR_SERVICE_ID} --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
```

## Delete a service

To delete a service, send this call:
```
curl -X DELETE https://gateway.kyma.local:$NODE_PORT/ec-default/v1/metadata/services/{YOUR_SERVICE_ID} --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
```

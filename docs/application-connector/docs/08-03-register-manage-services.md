---
title: Register a service
type: Tutorials
---

This guide shows you how to register a service of your external solution in Kyma.

## Prerequisites

Valid certificate signed by the Kyma Certificate Authority.

## Register a service

1. To register a service with a Basic Authentication-secured API, follow this template to prepare the request body:
  >**NOTE:** Follow [this](#tutorials-register-a-secured-api) tutorial to learn how to register APIs secured with different security schemes or protected against cross-site request forgery (CSRF) attacks.

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
      "spec": {},
      "credentials": {
        "basic": {
          "username": "{USERNAME}",
          "password": "{PASSWORD}"
        }
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

2. Include the request body you prepared in the following call to register a service:

  - For a cluster deployment:
    ```
    curl -X POST -d '{YOUR_REQUEST_BODY}' https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
    ```

  - For a local deployment:
    ```
    curl -X POST -d '{YOUR_REQUEST_BODY}' https://gateway.kyma.local:$NODE_PORT/{APP_NAME}/v1/metadata/services --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
    ```

A successful response returns the ID of the registered service:
```
{"id":"{YOUR_SERVICE_ID}"}
```

### Check the details of a registered service

  - For a cluster deployment:
    ```
    curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{YOUR_SERVICE_ID} --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
    ```

  - For a local deployment:
    ```
    curl https://gateway.kyma.local:{NODE_PORT}/{APP_NAME}/v1/metadata/services/{YOUR_SERVICE_ID} --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
    ```


## Register API with specification URL

Application Registry allows you to pass API specification in form of specification URL.

To register API with specification URL replace `api.spec` with `api.specificationUrl`. Be aware that `api.spec` has higher priority than `api.specificationUrl`.

```
"api": {
  "targetUrl": "https://services.odata.org/OData/OData.svc",
  "specificationUrl": "https://services.odata.org/OData/OData.svc/$metadata",
  "credentials": {
    "basic": {
      "username": "{USERNAME}",
      "password": "{PASSWORD}"
    }
}
```

The Application Registry will fetch the specification from provided URL but it will not use any credentials, therefor the endpoint can not be secured by any authentication mechanism.

>**NOTE:** Fetching specification from URL is supported only for API Spec. It can not be done for Events and Documentation.


## Registering OData API

If no `api.spec` or `api.specificationUrl` are specified and `api.type` is set to `OData`, the Application Registry will try to fetch the specification from the target URL with the `$metadata` path.

For service with the following api:
```
"api": {
  "targetUrl": "https://services.odata.org/OData/OData.svc",
  "apiType": "OData"
  "credentials": {
    "basic": {
      "username": "{USERNAME}",
      "password": "{PASSWORD}"
    }
}
```

Application Registry will try to fetch API Spec from `https://services.odata.org/OData/OData.svc/$metadata`.
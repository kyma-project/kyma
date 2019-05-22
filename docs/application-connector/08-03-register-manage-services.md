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
      "queryParameters": {
        "param": ["bar"]
      },
      "headers": {
        "custom-header": ["foo"]
      },
      "credentials": {
        "basic": {
          "username": "{USERNAME}",
          "password": "{PASSWORD}"
        }
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
    curl -X POST -d '{YOUR_REQUEST_BODY}' https://gateway.kyma.local:{NODE_PORT}/{APP_NAME}/v1/metadata/services --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
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


## Register API with a specification URL

The Application Registry allows you to pass API specifications in a form of specification URLs.

To register API with specification URL, replace `api.spec` with `api.specificationUrl`. 

>**NOTE:** If both api.spec and api.specificationUrl are provided, api.spec will be used due to its higher priority.

See the example of the API part of the request body with specification URL:
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

The Application Registry will fetch the specification from provided URL but it will not use any credentials, therefore the endpoint can not be secured by any authentication mechanism.

>**NOTE:** Fetching specification from a URL is supported only for APIs. Fetching specifications for Events or documentation is not supported.


## Register the OData API

If the **api.spec** or **api.specificationUrl** parameters are not specified and the **api.type** parameter is set to `OData`, the Application Registry will try to fetch specification from the target URL with the `$metadata` path.

For example, for the service with the following API, the Application Registry will try to fetch API specification from `https://services.odata.org/OData/OData.svc/$metadata`.
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

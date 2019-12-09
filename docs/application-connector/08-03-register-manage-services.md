---
title: Register a service
type: Tutorials
---

This guide shows you how to register a service of your external solution in Kyma.

## Prerequisites

- A valid certificate signed by the Kyma Certificate Authority

## Register a service

1. To register a service with a Basic Authentication-secured API, follow this template to prepare the request body:

   >**NOTE:** Follow [this](#tutorials-register-a-secured-api) tutorial to learn how to register APIs secured with different security schemes or protected against cross-site request forgery (CSRF) attacks.

   ```json
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
       "requestParameters": {
         "queryParameters": {
           "param": ["bar"]
         },
         "headers": {
           "custom-header": ["foo"]
         }
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
   
   ```bash
   curl -X POST -d '{YOUR_REQUEST_BODY}' https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
   ```

   A successful response returns the ID of the registered service:

   ```json
   {"id":"{YOUR_SERVICE_ID}"}
   ```

### Check the details of a registered service
   
```bash
curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{YOUR_SERVICE_ID} --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
```

## Register an API with a specification URL

The Application Registry allows you to pass API specifications in a form of specification URLs.

To register an API with a specification URL, replace `api.spec` with `api.specificationUrl`. 

>**NOTE:** If both api.spec and api.specificationUrl are provided, api.spec will be used due to its higher priority.

See the example of the API part of the request body with a specification URL:

```json
"api": {
  "targetUrl": "https://services.odata.org/OData/OData.svc",
  "specificationUrl": "https://services.odata.org/OData/OData.svc/$metadata",
  "credentials": {
    "basic": {
      "username": "{USERNAME}",
      "password": "{PASSWORD}"
    }
  }
}
```

>**NOTE:** Fetching a specification from a URL is supported only for APIs. Fetching specifications for events or documentation is not supported.

## Register an API with a secured specification URL

The Application Registry allows you to register an API with a secured specification URL. The supported authentication methods are [Basic Authentication](https://tools.ietf.org/html/rfc7617) and [OAuth](https://tools.ietf.org/html/rfc6750). You can specify only one type of authentication for an API.

### Register an API with a Basic Authentication-secured specification URL

To register an API with a specification URL secured with Basic Authentication, add a `specificationCredentials.basic` object to the `api` section of the service registration request body. You must include these fields:

| Field   |  Description |
|----------|------|
| **username** | Basic Authorization username |
| **password** | Basic Authorization password |

This is an example of the `api` section of the request body for an API with a specification URL secured with Basic Authentication:

```json
    "api": {
        "targetUrl": "https://sampleapi.targeturl/v1",
        "specificationUrl": "https://sampleapi.spec/v1",
        "specificationCredentials": {
            "basic": {
                "username": "{USERNAME}",
                "password": "{PASSWORD}"
            },
        }  
    }
```

### Register an API with an OAuth-secured specification URL

To register an API with a specification URL secured with OAuth, add a `specificationCredentials.oauth` object to the `api` section of the service registration request body. Include these fields in the request body:

| Field   |  Description |
|----------|------|
| **url** |  OAuth token exchange endpoint of the service |
| **clientId** | OAuth client ID |
| **clientSecret** | OAuth client Secret |
| **requestParameters.headers** | Additional request headers (optional)|   
| **requestParameters.queryParameters** | Additional query Parameters (optional)| 

This is an example of the `api` section of the request body for an API with a specification URL secured with OAuth:

```json
    "api": {
        "targetUrl": "https://sampleapi.targeturl/v1",
        "specificationUrl": "https://sampleapi.spec/v1",
        "specificationCredentials": {
            "oauth": {
                "url": "https://sampleapi.targeturl/authorizationserver/oauth/token",
                "clientId": "{CLIENT_ID}",
                "clientSecret": "{CLIENT_SECRET}",
                "requestParameters" : {
                     "headers": {
                         "{CUSTOM_HEADER_NAME}" : ["{CUSTOM_HEADER_VALUE}"]
                     },
                     "queryParameters":  {
                         "{CUSTOM_QUERY_PARAMETER_NAME}" : ["{CUSTOM_QUERY_PARAMETER_VALUE}"]
                     }
                }               
            }
        }  
    }
```

## Use custom headers and query parameters for fetching API specification from URL 

You can specify additional headers and query parameters to inject to requests made to the specification URL.

>**NOTE:** These headers and query parameters are used only for requests for fetching an API specification and are not stored in the system.

To register an API with a specification URL that requires specific custom headers and query parameters, add the `specificationRequestParameters.headers` and `specificationRequestParameters.queryParameters` objects to the `api` section of the service registration request body.

```json
    "api": {
        "targetUrl": "https://sampleapi.targeturl/v1",
        "specificationUrl": "https://sampleapi.spec/v1",
        "specificationRequestParameters": {
            "headers": {
                "{CUSTOM_HEADER_NAME}": ["{CUSTOM_HEADER_VALUE}"]
            },
            "queryParameters": {
                "{CUSTOM_QUERY_PARAMETER_NAME}" : ["{CUSTOM_QUERY_PARAMETER_VALUE}"]
            },
        }
        "credentials": {
            "basic": {
                "username": "{USERNAME}",
                "password": "{PASSWORD}"
            },
        }
    }
```

## Register an OData API

If the **api.spec** or **api.specificationUrl** parameters are not specified and the **api.type** parameter is set to `OData`, the Application Registry will try to fetch the specification from the target URL with the `$metadata` path.

For example, for the service with the following API, the Application Registry will try to fetch the API specification from `https://services.odata.org/OData/OData.svc/$metadata`.

```json
"api": {
  "targetUrl": "https://services.odata.org/OData/OData.svc",
  "apiType": "OData"
  "credentials": {
    "basic": {
      "username": "{USERNAME}",
      "password": "{PASSWORD}"
    }
  }
}
```

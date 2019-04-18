---
title: Asset Upload Service
type: Details
---

The Asset Upload Service is an HTTP server that exposes the file upload functionality for Minio. It contains a simple HTTP endpoint which accepts `multipart/form-data` forms. It can upload files to the private and public system buckets.

The main purpose of the service is to provide a solution for hosting static files for components that use the Asset Store, such as the Application Connector.
You can also use the Asset Upload Service for development purposes to host files for the Asset Store, without the need to rely on external providers.

## System buckets

The Asset Upload Service creates two system buckets, `system-private-{generated-suffix}` and `system-public-{generated-suffix}`, where `{generated-suffix}` is a Unix nano timestamp in the 32-base number system. The public bucket has a read-only policy specified.

To enable the service scaling and to maintain the bucket configuration data between the application restarts, the Asset Upload Service stores its configuration in the `assetstore-asset-upload-service` ConfigMap.

Once you upload the files, system buckets store them permanently. There is no policy to clean system buckets periodically.

The diagram describes the Asset Upload Service flow:


![](./assets/asset-upload-service.svg)


## Use the service outside the Kyma cluster

You can expose the service for development purposes. To use the Asset Upload Service on a local machine, run the following command:

```bash
kubectl port-forward deployment/assetstore-asset-upload-service 3000:3000 -n kyma-system
```

You can access the service on port `3000`.


### Upload files

To upload files, send the multipart form **POST** request to the `/v1/upload` endpoint. The endpoint recognizes the following field names:

- `private` that is an array of files to upload to a private system bucket.
- `public` that is an array of files to upload to a public system bucket.
- `directory` that is an optional directory for storing the uploaded files. If you do not specify it, the service creates a directory with a random name. If the directory and files already exist, the service overwrites them.

To do the multipart request using `curl`, run the following command:

```bash
curl -v -F directory='example' -F private=@sample.md -F private=@text-file.md -F public=@archive.zip http://localhost:3000/v1/upload
```

The result is as follows:

```json
{
   "uploadedFiles": [
      {
         "fileName": "text-file.md",
         "remotePath": "https://minio.kyma.local/private-1b0sjap35m9o0/example/text-file.md",
         "bucket": "private-1b0sjap35m9o0",
         "size": 212
      },
      {
         "fileName": "archive.zip",
         "remotePath": "https://minio.kyma.local/public-1b0sjaq6t6jr8/example/archive.zip",
         "bucket": "public-1b0sjaq6t6jr8",
         "size": 630
      },
      {
         "fileName": "sample.md",
         "remotePath": "https://minio.kyma.local/private-1b0sjap35m9o0/example/sample.md",
         "bucket": "private-1b0sjap35m9o0",
         "size": 4414
      }
   ]
}
```

See the [OpenAPI specification](./assets/asset-upload-service-openapi.yaml) for the full API documentation. You can use the [Swagger Editor](https://editor.swagger.io) to preview and test the API service.

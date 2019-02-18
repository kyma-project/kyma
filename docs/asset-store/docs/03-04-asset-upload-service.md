---
title: Asset Upload Service
type: Details
---

Asset Upload Service is a HTTP server, that exposes file upload functionality for Minio. It contains a simple HTTP endpoint, which accepts multiple files as form. It can upload files for two system buckets: private and public one, with read-only policy set.

The main purpose of the service is to provide upload capability for Application Connector. Some files, such as Async API spec, can be inaccessible from inside the cluster. To enable document upload with Asset Controller, they need to be uploaded in a temporary space.

Asset Upload Service can be used for development purposes. You can use this service to temporary host files for Asset Store, without need to use external providers.

## System buckets 

The Asset Upload Service creates two system buckets: `system-private-{generated-suffix}` and `system-public-{generated-suffix}`, where `{generated-suffix}` is Unix nano timestamp in 32-base number system. Public bucket has read-only policy specified.
 
To enable scaling and to keep bucket configuration data between application restarts, the Asset Upload Service stores its configuration in `assetstore-asset-upload-service` ConfigMap.

## Usage outside Kyma cluster

You can expose the service for development purposes. In order to use Asset Upload Service on local machine, run the following command:

```bash
kubectl port-forward deployment/assetstore-asset-upload-service 3000:3000 -n kyma-system
```

The service will be accessible on 3000 port on local machine.


### Upload files

To upload files, send a Multipart form POST request to `/upload` endpoint. The endpoint recognizes the following field names:

- `private` - array of files, which should be uploaded to private system bucket.  
- `private` - array of files, which should be uploaded to public read-only system bucket.  
- `directory` - optional directory, where the uploaded files are put. If it is not specified, it will be randomized.

To do the multipart request using `curl`, run the following command in this repository:

```bash
curl -v -F directory='example' -F private=@sample.md -F private=@text-file.md -F public=@archive.zip http://localhost:3000/upload
```

The result is:

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

You can use 
---
title: Front Matter Service
type: Details
---

The Front Matter Service is an HTTP server that exposes the functionality for extracting metadata from files. It contains a simple HTTP endpoint which accepts `multipart/form-data` forms. The service extracts front matter YAML metadata from text files of all extensions.

The main purpose of the service is to provide metadata extraction for Rafter controllers. That's why it is only available inside the cluster. To use it, define `metadataWebhookService` in Asset and ClusterAsset custom resources.

## Front matter metadata

Front matter YAML metadata are YAML properties added at the beginning of a file, between `---` lines. The following snippet represents an exemplary Markdown file with metadata specified:

```markdown

---
title: Example document title
description: Example page description
order: 3
array:
 - foo
 - bar
---

## Lorem ipsum
Dolores sit amet

```

## Use the service outside the Kyma cluster

You can expose the service for development purposes. To use the Front Matter Service on a local machine, run the following command:

```bash
kubectl port-forward deployment/rafter-front-matter-service 3000:3000 -n kyma-system
```

You can access the service on port `3000`.

### Metadata files

To extract metadata from files, send the multipart form **POST** request to the `/v1/extract` endpoint. Specify the relative or absolute path to the file as a field name.
To do the multipart request using `curl`, run the following command:

```bash
curl -v -F foo/foo.md=@foo.md -F bar/bar.yaml=@bar.yaml http://localhost:3000/v1/extract
```

The result is as follows:

```json
{
  "data": [
    {
      "filePath": "foo/foo.md",
      "metadata": {
        "no": 3,
        "title": "Access logs",
        "type": "Details"
      }
    },
    {
      "filePath": "bar/bar.yaml",
      "metadata": {
        "number": 9,
        "title": "Hello world",
        "url": "https://kyma-project.io"
      }
    }
  ]
}
```

See the [OpenAPI specification](./assets/front-matter-service-openapi.yaml) for the full API documentation.

---
title: Create a bundles repository
type: Details
---

The repository in which you create your own bundles must have a specific format so that the Helm Broker can fetch bundles from it. Your remote bundle repository must include the following resources:

```
sample-bundle-repository
  ├── {your-bundle}                # A directory for each bundle version defined in the index.yaml file
  └── index.yaml                   # A file which defines available bundles
```

Follow the `{bundle_name}-{bundle_version}` convention to name your bundles. In the `index.yaml` file, provide an entry for every single bundle from your bundles repository. The `index.yaml` file must have the following structure:

```
apiVersion: v1
entries:
  {bundle_name}:
    - name: {bundle_name}
      description: {bundle_description}
      version: {bundle_version}
```

See the example:

```yaml
apiVersion: v1
entries:
  redis:
    - name: redis
      description: Redis service
      version: 0.0.1
```

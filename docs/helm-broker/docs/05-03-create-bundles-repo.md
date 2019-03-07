---
title: Create a bundle repository
type: Configuration
---

The repository in which you create your own bundles must have a specific format so that the Helm Broker can fetch bundles from it. Your remote bundle repository must include the following resources:

```
sample-bundle-repository
  ├── {your-bundle}                # A directory for each bundle version defined in the index.yaml file
  ├── index.yaml                   # A file which defines available bundles
```

This `index.yaml` file must have the following structure:

```
apiVersion: v1
entries:
  {bundle_name}:
    - name: {bundle_name}
      description: {bundle_description}
      version: {bundle_version}
```

See the example `index.yaml` file for the Redis bundle:

```
apiVersion: v1
entries:
  redis:
    - name: redis
      description: Redis service
      version: 0.0.1
```



A `{bundle_name}-{bundle_version}.tgz` file for each bundle version defined in the `yaml` file. The `.tgz` file is an archive of your bundle's directory.


The Helm Broker fetches bundle definitions from HTTP servers defined in the `helm-repos-urls` ConfigMap.

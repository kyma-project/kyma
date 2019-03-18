---
title: Create a bundles repository
type: Details
---

The repository in which you create your own bundles must have a specific format so that the Helm Broker can fetch bundles from it. Your remote bundle repository must include the following resources:

```
sample-bundle-repository
  ├── {your-bundle}.tgz            # A directory for each bundle version defined in the index.yaml file
  └── index.yaml                   # A file which defines available bundles
```

Follow the `{bundle_name}-{bundle_version}` convention to name your bundles. Provide your bundles in the `.tgz` format. To learn how to create your own bundle, read [this](#details-create-a-bundle) document.

### index.yaml file

In the `index.yaml` file, provide an entry for every single bundle from your bundles repository. The `index.yaml` file must have the following structure:

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

Depending on your needs and preferences, you can create one or more `index.yaml` files with entries to your bundles. For example, you can have separate `index-testing.yaml` file for testing bundles, and `index-production.yaml` file for production bundles. However, you must include at least one file named `index.yaml` as the Helm Broker automatically searches for it when searching for `.../{path_to_your_bundle}/{bundle_version}/` URL.

>**CAUTION:** Do not provide entries for a given bundle in two separate `index.yaml` files. In such a case, the Helm Broker does not read this bundle at all.

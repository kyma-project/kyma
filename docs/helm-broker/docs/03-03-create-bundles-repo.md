---
title: Create a bundles repository
type: Details
---

The repository in which you create your own bundles must be an HTTPS server with a specific structure so that the Helm Broker can fetch bundles from it. Your remote bundle repository can contain many bundles, each one compressed to the `.tgz` format and defined in `index.yaml` files. Depending on your needs and preferences, you can create one or more `index.yaml` files to categorize your bundles. The repository structure looks as follows:

```
sample-bundle-repository
  ├── {bundle_x_name}-{bundle_x_version}.tgz         # A bundle compressed to a .tgz file
  ├── {bundle_y_name}-{bundle_y_version}.tgz        
  ├── ...                                      
  ├── index.yaml                                     # A file which defines available bundles
  ├── index-2.yaml                              
  └── ...                                                    
```

Read [this](https://github.com/kyma-project/bundles/blob/master/docs/getting-started.md) document to learn how to set up your own bundles repository which generates `.tgz` and `index.yaml` files, and expose them using an HTTPS server. See the example of the Kyma `bundles` repository [here](https://github.com/kyma-project/bundles/releases).

### {bundle_name}-{bundle_version}.tgz file

The `{bundle_name}-{bundle_version}.tgz` file is a compressed version of your bundle. To learn how to create your own bundle, read [this](#details-create-a-bundle) document.

>**TIP:** If you contribute to the [bundles](https://github.com/kyma-project/bundles/tree/master/bundles) repository, you do not have to compress your bundles as the system does it automatically.

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

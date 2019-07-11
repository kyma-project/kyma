---
title: Create addons repository
type: Details
---

The repository in which you create your own addons must be an HTTPS server with a specific structure so that the Helm Broker can fetch addons from it. Your remote addons repository can contain many addons, each one compressed to the `.tgz` format and defined in `index.yaml` files. Depending on your needs and preferences, you can create one or more `index.yaml` files to categorize your addons. The repository structure looks as follows:

```
sample-addon-repository
  ├── {addon_x_name}-{addon_x_version}.tgz         # An addon compressed to a .tgz file
  ├── {addon_y_name}-{addon_y_version}.tgz        
  ├── ...                                      
  ├── index.yaml                                     # A file which defines available addons
  ├── index-2.yaml                              
  └── ...                                                    
```

Read [this](https://github.com/kyma-project/bundles/blob/master/docs/getting-started.md) document to learn how to set up your own addons repository which generates `.tgz` and `index.yaml` files, and expose them using an HTTPS server. See the example of the Kyma `bundles` repository [here](https://github.com/kyma-project/bundles/releases).

### {addon_name}-{addon_version}.tgz file

The `{addon_name}-{addon_version}.tgz` file is a compressed version of your addon. To learn how to create your own addon, read [this](#details-create-addons) document.

>**TIP:** If you contribute to the [bundles](https://github.com/kyma-project/bundles/tree/master/bundles) repository, you do not have to compress your addons as the system does it automatically.

### index.yaml file

In the `index.yaml` file, provide an entry for every single addon from your addons repository. The `index.yaml` file must have the following structure:

```
apiVersion: v1
entries:
  {addon_name}:
    - name: {addon_name}
      description: {addon_description}
      version: {addon_version}
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

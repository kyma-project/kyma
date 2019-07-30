---
title: Create addons repository
type: Details
---

The repository in which you create your own addons must have a specific structure and be exposed as a server so that the Helm Broker can fetch addons from it. Your remote addons repository can contain many addons defined in index `.yaml` files. Depending on your needs and preferences, you can create one or more `index.yaml` files to categorize your addons. In the `index.yaml` file, provide an entry for every single addon from your addons repository. The `index.yaml` file must have the following structure:
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

You must place your addons in the same directory where the `index.yaml` file is stored. The Helm Broker supports the following servers to expose your addons:

<div tabs>
  <details>
  <summary>
  HTTP/HTTPS
  </summary>

>**NOTE:** The HTTP protocol is supported only in `DevelopMode`. To learn more, read [this](#details-registration-rules-using-http-urls) document.

If you want to expose your addons using an HTTPS server, you must compress your addons to `.tgz` files. The repository structure looks as follows:
```
sample-addon-repository
  ├── {addon_x_name}-{addon_x_version}.tgz           # An addon compressed to a .tgz file
  ├── {addon_y_name}-{addon_y_version}.tgz        
  ├── ...                                      
  ├── index.yaml                                     # A file which defines available addons
  ├── index-2.yaml                              
  └── ...                                                    
```

>**TIP:** If you contribute to the [addons](https://github.com/kyma-project/addons/tree/master/addons) repository, you do not have to compress your addons as the system does it automatically.

See the example of the Kyma `addons` repository [here](https://github.com/kyma-project/addons/releases).

  </details>
  <details>
  <summary>
  Git
  </summary>

If you want to expose your addons using Git, place your addons in addons directories. The repository structure looks as follows:
```
sample-addon-repository
  ├── {addon_x_name}-{addon_x_version}               # An addon directory
  ├── {addon_y_name}-{addon_y_version}        
  ├── ...                                      
  ├── index.yaml                                     # A file which defines available addons
  ├── index-2.yaml                              
  └── ...                                                    
```
You can specify Git repositories by adding the special `git::` prefix to the URL address. After this prefix, you can provide any valid Git URL with one of the protocols supported by Git. These are the allowed git URLs:
- git::https://github.com/kyma-project/addons.git//addons/index-testing.yaml?ref={branch,commit_sha,tag}
- github.com/kyma-project/addons//addons/index.yaml?ref={branch,commit_sha,tag}
- bitbucket.org/kyma-project/addons//addons/index.yaml

>**NOTE:** For now, SSH protocol is not supported.

See the example of the Kyma `addons` repository [here](https://github.com/kyma-project/addons/tree/master/addons).

  </details>
</div>

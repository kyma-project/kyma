---
title: Create addons repository
type: Details
---

The repository in which you create your own addons must contain at least one `index.yaml` file and have a specific structure, depending on the type of server that exposes your addons.

## The index yaml file

Your remote addons repository can contain many addons defined in index files. Depending on your needs and preferences, you can create one or more index files to categorize your addons. In the index file, provide an entry for every single addon from your addons repository. The index file must have the following structure:
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

>**NOTE:** You must place your addons in the same directory where the `index.yaml` file is stored.

## Supported protocols

Expose your addons repository so that you can provide URLs in the [AddonsConfiguration](#custom-resource-addonsconfiguration) (AC) and [ClusterAddonsConfiguration](#custom-resource-clusteraddonsconfiguration) (CAC) custom resources. The Helm Broker supports exposing addons using the following protocols:

<div tabs>
  <details>
  <summary>
  HTTP/HTTPS
  </summary>

>**NOTE:** The HTTP protocol is supported only in `DevelopMode`. To learn more, read [this](#details-registration-rules-using-http-urls) document.

If you want to use an HTTP or HTTPS server, you must compress your addons to `.tgz` files. The repository structure looks as follows:
```
sample-addon-repository
  ├── {addon_x_name}-{addon_x_version}.tgz           # An addon compressed to a .tgz file
  ├── {addon_y_name}-{addon_y_version}.tgz        
  ├── ...                                      
  ├── index.yaml                                     # A file which defines available addons
  ├── index-2.yaml                              
  └── ...                                                    
```

See the example of the Kyma `addons` repository [here](https://github.com/kyma-project/addons/releases).

>**TIP:** If you contribute to the Kyma [`addons`](https://github.com/kyma-project/addons/tree/master/addons) repository, you do not have to compress your addons as the system does it automatically.

These are the allowed addon repository URLs provided in CAC or AC custom resources for HTTP or HTTPS servers:
```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  name: addons-cfg-sample
spec:
  repositories:
    # HTTPS protocol
    - url: "https://github.com/kyma-project/addons/releases/download/latest/index.yaml"
    # HTTP protocol
    - url: "http://github.com/kyma-project/addons/releases/download/latest/index.yaml"
```

  </details>
  <details>
  <summary>
  Git
  </summary>

If you want to use Git, place your addons directly in addons directories. The repository structure looks as follows:
```
sample-addon-repository
  ├── {addon_x_name}-{addon_x_version}               # An addon directory
  ├── {addon_y_name}-{addon_y_version}        
  ├── ...                                      
  ├── index.yaml                                     # A file which defines available addons
  ├── index-2.yaml                              
  └── ...                                                    
```

See the example of the Kyma `addons` repository [here](https://github.com/kyma-project/addons/tree/master/addons).

> **NOTE:** The amount of memory and storage size determine the maximum size of your addons repository. These limits are set in the
[Helm Broker chart](https://kyma-project.io/docs/components/helm-broker/#configuration-helm-broker-chart).

You can specify a Git repository URL by adding a special `git::` prefix to the URL address. After this prefix, provide any valid Git URL with one of the protocols supported by Git. In the URL, you can specify a branch, commit, or tag version. You can also add the `depth` query parameter with a number that specifies the last revision you want to clone from the repository.

>**NOTE:** If you use `depth` together with `ref`, make sure that `depth` number is big enough to clone a proper reference. For example, if you have `depth=1` and `ref` set to a commit from the distant past, the URL will not work as you clone only the first commit from the `master` branch and there is no option to do the checkout.

These are the allowed addon repository URLs provided in CAC or AC custom resources for Git:
```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  name: addons-cfg-sample
spec:
  repositories:
    # Git HTTPS protocol with a path to index.yaml
    - url: "git::https://github.com/kyma-project/addons.git//addons/index.yaml"
    # Git HTTPS protocol with a path to index.yaml of a specified version and a depth query parameter
    - url: "git::https://github.com/kyma-project/addons.git//addons/index.yaml?ref=1.2.0&depth=3"
    # github.com URL with no prefix. It is automatically interpreted as a Git repository source.
    - url: "github.com/kyma-project/addons//addons/index.yaml"
    # bitbucket.org URL with no prefix. It is automatically interpreted as a Git repository source.
    - url: "bitbucket.org/kyma-project/addons//addons/index.yaml"
```

  </details>
  <details>
  <summary>
  Mercurial
  </summary>

If you want to use Mercurial (hg), place your addons directly in addons directories. The repository structure looks as follows:
```
sample-addon-repository
  ├── {addon_x_name}-{addon_x_version}               # An addon directory
  ├── {addon_y_name}-{addon_y_version}        
  ├── ...                                      
  ├── index.yaml                                     # A file which defines available addons
  ├── index-2.yaml                              
  └── ...                                                    
```
> **NOTE:** The amount of memory and storage size determine the maximum size of your addons repository. These limits are set in the
[Helm Broker chart](https://kyma-project.io/docs/components/helm-broker/#configuration-helm-broker-chart).

You can specify a Mercurial repository URL by adding a special `hg::` prefix to the URL address. After this prefix, provide a valid Mercurial URL with one of the supported protocols. In the URL, you can specify a revision to checkout.

These are the allowed addon repository URLs provided in CAC or AC custom resources for Mercurial:
```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  name: addons-cfg-sample
spec:
  repositories:
    # Mercurial HTTPS protocol with a path to index.yaml
    - url: "hg::https://hg.osdn.net/view/project-name/repo-name//index.yaml"
    # Mercurial HTTPS protocol with a path to index.yaml and a revision
    - url: "hg::https://hg.osdn.net/view/project-name/repo-name//index.yaml?rev=e67e535230d4eded318b30967e32397872e53af1"
```

  </details>

</div>

## Supported protocols with authentication

You can provide authentication as part of the URL in your AddonsConfiguration and ClusterAddonsConfiguration custom resources. Instead of using sensitive information directly in the URL, put it in the Secret resource and take advantage of templating. In your repository URL, use placeholders which refer to keys in the Secret. See the following example:

```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  name: addons-cfg-sample
spec:
  repositories:
    - url: "https://{host}/{project}/addons/index.yaml"
      secretRef:
        name: data
---
apiVersion: v1
kind: Secret
metadata:
  name: data
type: Opaque
stringData:
  host: "github.com"
  project: "kyma-project/addons"       
```
The URL renders as follows:
```
https://github.com/kyma-project/addons/addons/index.yaml
```

The Helm Broker supports authentication with these protocols:

<div tabs>
  <details>
  <summary>
  HTTP/HTTPS
  </summary>

To secure your addons repository with basic authentication credentials, create a Secret resource which contains credentials, and reference it in the repository URL definition using templating. Follow these steps:

1. Create a Secret:

   ```bash
   kubectl create secret generic auth -n stage --from-literal=username=admin --from-literal=password=secretPassword
   ```

2. In your repository URL, precede the hostname with the `username:password@` section:

   ```
   https://admin:secretPassword@repository.addons.com/index.yaml
   ```

3. Define a ClusterAddonsConfiguration or AddonsConfiguration custom resource:

   ```yaml
   apiVersion: addons.kyma-project.io/v1alpha1
   kind: ClusterAddonsConfiguration
   metadata:
     name: addons-cfg-sample
   spec:
     repositories:
       # HTTPS protocol with basic authorization provided
       - url: "https://{username}:{password}@repository.addons.com/index.yaml"
         secretRef:
           name: auth
           namespace: stage
   ```

  </details>
  <details>
  <summary>
  Git SSH
  </summary>

  The Git SSH protocol requires an SSH key to authenticate with your repository. Setting SSH keys differs among hosting providers. Read [this](https://help.github.com/en/articles/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent#generating-a-new-ssh-key) document to learn how to generate a new SSH key in the GitHub service.
  >**NOTE:** The Git SSH private key must be base64-encoded.

  Follow these steps to secure your addons repository with basic authentication credentials:

  1. Run this command to encode your private key:
  ```bash
    base64 -b -i {path_to_id_rsa} -o id_rsa-encoded
  ```
  > **NOTE:** Do not secure your private SSH key with a passphrase.

  2. Create a corresponding Secret resource:
  ```bash
  kubectl create secret generic auth -n stage --from-file=key=id_rsa-encoded
  ```

  3. Define a URL with the required private SSH key option:
  ```yaml
  apiVersion: addons.kyma-project.io/v1alpha1
  kind: ClusterAddonsConfiguration
  metadata:
    name: addons-cfg-sample
  spec:
    repositories:
      # Git SSH protocol with the reference to the Secret that contains base64-encoded SSH private key
      - url: "git::ssh://git@github.com/kyma-project/private-addons.git//addons/index.yaml?sshkey={key}"
        secretRef:
          name: auth
          namespace: stage
  ```

  </details>
  <details>
  <summary>
  S3
  </summary>

  The S3 protocol requires a key and a secret to authenticate with your bucket. To get a key and a secret, log in to the AWS [console](https://console.aws.amazon.com),
  select **My Security Credentials**, and go to the **Access keys** tab where you can create a new access key.

  > **NOTE:** For more information about security credentials and access key, read [this](https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html) document.

  Follow these steps to secure your S3 addons repository with basic authentication credentials:

  1. Create a Secret resource with S3 credentials:
  ```bash
    echo -n 'AWS_KEY' > ./key.txt
    echo -n 'AWS_SECRET' > ./secret.txt

    kubectl create secret generic aws-auth -n default --from-file=aws_key=./key.txt --from-file=aws_secret=./secret.txt
  ```

  2. Define a URL with the required fields:
  ```yaml
    apiVersion: addons.kyma-project.io/v1alpha1
    kind: ClusterAddonsConfiguration
    metadata:
      name: addons-cfg-sample
    spec:
      repositories:
        - url: "s3::http://s3-region.amazonaws.com/addon-test/addons//index.yaml?aws_access_key_id={aws_key}&aws_access_key_secret={aws_secret}"
          secretRef:
            name: aws-auth
            namespace: default
  ```

  </details>
</div>  

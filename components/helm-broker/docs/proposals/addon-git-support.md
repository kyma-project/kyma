# Support fetching addons from Git repository

Created on 2019-07-24 by Mateusz Szostok ([@mszostok](https://github.com/mszostok)).

This document describes options to extend the ClusterAddonsConfiguration and AddonsConfiguration custom resources (CRs) so that they support fetching addons from any Git repository.

## Motivation

Helm Broker fetches addons listed in the `index.yaml` file and exposed as remote HTTPS servers, which means that only zipped addons can be fetched. This solution generates problems for end-users, as it's not always easy to have a dedicated HTTPS server that serves zipped addons. The most common scenario is to have a Git repository with unzipped addons and index yaml files. This document describes all possible solutions for achieving that goal.

## Accepted solution

You can specify Git repositories by adding the special `git::` prefix to the URL address. After this prefix, you can provide any valid Git URL with one of the protocols supported by Git. This solution is inspired by the [Terraform implementation](https://www.terraform.io/docs/modules/sources.html#github). Using the [hashicorp go-getter](https://github.com/hashicorp/go-getter) allows us to easily add new supported protocols, such as Mercurial, S3, GCS, etc.

```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  name: addons-cfg-sample
spec:
  repositories:
    # Use Git SSH protocol with a path to index.yaml
    - url: "git::git@github.com:kyma-project/addons.git//index.yaml"
    # Use Git HTTPS protocol with a path to index.yaml 
    - url: "git::https://github.com/kyma-project/addons.git//addons/index.yaml"
    # Use Git HTTPS protocol with a path to index.yaml and branch/tag version 
    - url: "git::https://github.com/kyma-project/addons.git//addons/index.yaml?ref=1.2.0"
    # Use unprefixed github.com URLs. They are automatically interpreted as Git repository sources. 
    - url: "github.com/kyma-project/addons//addons/index.yaml?ref=1.2.0"
    # Use HTTPS protocol (server which serves static content) defined the way it was implemented so far.
    - url: "https://github.com/kyma-project/addons/releases/download/latest/index.yaml"
``` 


<details><summary>Pros</summary>
<p>

- No UI and Console Backend Service changes
- No migration job needed 
- Support for all Git repositories
- The Git getter is already available in [hashicorp](https://github.com/hashicorp/go-getter#git-git). We only need to create a validator that checks if the file is provided (in hashicorp it is optional). 

>**NOTE:** This implementation uses the Git binary. It check-outs the full repository but depth can be specified.

</p>
</details>

<details><summary>Cons</summary>
<p>

- User must know the new pattern, which is not obvious.
- Link is not clickable. You cannot click the link directly in UI and go to a given repository. 

</p>
</details>

Mitigation:
Link is not clickable, however, under the **status** entry of both CRs, there is information about the content fetched from index.yaml, so going directly to the specified file is not critical. We only need to change the UI to display the **status** entry as well. 

> **NOTE:** The internal implementation should use the [hashicorp go-getter](https://github.com/hashicorp/go-getter).

> **NOTE:** In the first phase, the HTTPS and Git protocols are officially supported and tested. All other protocols, such as S3 or GCS, are out of scope.

## Rejected solutions

This section contains the initial propositions which were rejected.

### Solution 1: Extend API with the **type** field 

Git repositories can be specified by choosing `type: git` and providing the direct address to a file in the Git repository.
The link is exactly the same as the one in the browser URL bar. URL could be different between different Git hosting.

```
- GitHub: https://github.com/kyma-project/addons/blob/master/addons/index.yaml
- GitLab: https://gitlab.com/kyma-project/addons/blob/master/addons/index.yaml
- Bitbucket: https://bitbucket.org/kyma-project/addons/src/master/addons/index.yaml
```


```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  name: addons-cfg-sample
spec:
  repositories:
    - type: git
      url: https://github.com/kyma-project/addons/blob/master/addons/index-testing.yaml
``` 

<details><summary>Pros</summary>
<p>

- Link is clickable. You can click the link directly in UI and go to index.yaml in the repository.

</p>
</details>

<details><summary>Cons</summary>
<p>

- Migration job needed
- Changes in UI and Console Backend Service needed
- Support only for Git hosting and not directly to all Git repositories. We need to create our parser. We should also mention in the documentation that repositories other than GitHub, GitLab, and BitBucket may not work.
  

</p>
</details>

### Solution 2: Extend API with the **type**, **ref**, **url**, **indexPath** fields

Git repositories can be specified by choosing `type: git` and providing rest information in dedicated fields.

```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  name: addons-cfg-sample
spec:
  repositories:
    - type: git
      ref: master # branch/tag/commit
      url: git@github.com:kyma-project/bundles.git # all git protocol supported
      indexPath: bundles/bundles/index.yaml
``` 

<details><summary>Pros</summary>
<p>

- Support all Git repositories
- API is self-describing

</p>
</details>

<details><summary>Cons</summary>
<p>

- Migration job needed
- Changes in UI and Console Backend Service needed
- Link is not clickable. You cannot click the link directly in UI and go to a given repository.

</p>
</details>


## Future implementations

We consider adding support for authorization in the future. 

### Support for authorization

In the future, we want to allow you to specify the reference to a Secret in which you store credentials. In URL, you must specify query parameters with key-value pairs. As a value, set the name of the key from a Secret which contains token data. Helm Broker fetches ClusterAddonsConfigurations and renders URLs by replacing the **{KEY_NAME_FROM_SECRET_PLACEHOLDER}** with proper values extracted from Secrets referenced by the **authSecretRef** field.

```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  name: addons-cfg-sample
spec:
  repositories:
    - url: "s3::http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id={KEY_NAME_FROM_SECRET_PLACEHOLDER}&aws_access_key_secret={KEY_NAME_FROM_SECRET_PLACEHOLDER}"
      authSecretRef:
      	name: abc
        namespace: xyz
```

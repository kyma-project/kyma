# Support for loading Addons from Git repository

Created on 2019-07-24 by Mateusz Szostok ([@mszostok](https://github.com/mszostok)).

This document describes options for extending the ClusterAddonsConfiguration/AddonsConfiguration for supporting loading Addons from the Git repository.

## Motivation

Helm Broker fetches addons listed in the `index.yaml` file from remote HTTPS servers. Only zipped bundles can be fetched. This solution generates problems for end-user as it's not easy to always have dedicated HTTPS server which serves zipped bundles. A most common scenario is to have a git repository with unzipped bundles and index YAMLs. This document describes possible and accepted solutions for achieving that goal.

## Accepted solution

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

- Support all git repositories
- API is self-describing

</p>
</details>

<details><summary>Cons</summary>
<p>

- Migration job needed
- Changes in UI and console-backend-service needed
- Link is not "clickable". In UI you are not able to click directly the link and go to a given repository.

</p>
</details>

Mitigation:
- link is not "clickable" but in CRD under the status entry, we have full information about the content loaded from the index, so going directly to specified index YAML is no so critical. We need to just change the UI to also display a status entry. 
- migration job is already available, adding one step is trivial

> **NOTE:** The internal implementation should use the [hashicorp go-getter](https://github.com/hashicorp/go-getter).

> **NOTE:** The https and git protocols are officially supported and tested. All other protocols like S3, GCS, etc. are out of scope.


## Rejected solutions

This section contains the initial propositions which were rejected.
 
### Solution 1: Custom URL schema

Git repositories can be specified by prefixing the address with the special `git::` prefix. After this prefix, any valid Git URL can be specified to select one of the protocols supported by Git. Inspired by the [Terraform implementation](https://www.terraform.io/docs/modules/sources.html#github).

```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  name: addons-cfg-sample
spec:
  repositories:
    # Use SSH protocol with path to index YAML
    - url: "git@github.com:kyma-project/bundles.git//index.yaml"
    # Use HTTPS protocol with path to index YAML 
    - url: "git::https://github.com/kyma-project/bundles.git//bundles/index.yaml"
    # Use HTTPS protocol with path to index YAML and branch/tag version 
    - url: "git::https://github.com/kyma-project/bundles.git//bundles/index.yaml?ref=1.2.0"
``` 


<details><summary>Pros</summary>
<p>

- No UI and console-backend-service changes
- No migration job needed 
- Support all git repositories
- The git getter is already available from [hashicorp](https://github.com/hashicorp/go-getter#git-git). We need to only write validator for checking that file is always provided (in hashicorp is optional). 

  NOTE: Implementation is using the git binary. It check-outs the full repository but depth can be specified.

</p>
</details>

<details><summary>Cons</summary>
<p>

- User needs to know a new pattern, which is not obvious
- Link is not "clickable". In UI you are not able to click directly the link and go to a given repository. 

</p>
</details>

### Solution 2: Extend API with `type` field 

Git repositories can be specified by choosing `type: git` and providing the direct address to file in the git repository.
The link is exactly that which can be found in the browser URL bar. URL could be different between different git hosting.


- GitHub: `https://github.com/kyma-project/bundles/blob/master/bundles/index.yaml`
- GitLab: `https://gitlab.com/kyma-project/bundles/blob/master/bundles/index.yaml`
- Bitbucket: `https://bitbucket.org/kyma-project/bundles/src/master/bundles/index.yaml`


```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  name: addons-cfg-sample
spec:
  repositories:
    - type: git
      url: https://github.com/kyma-project/bundles/blob/master/bundles/index-testing.yaml
``` 

<details><summary>Pros</summary>
<p>

- Link is "clickable". In UI you are able to click directly the link and go to index in the repository.

</p>
</details>

<details><summary>Cons</summary>
<p>

- Migration job needed
- Changes in UI and console-backend-service needed
- Support only for git hosting and not directly to all git repositories. We need to write our own parser. 
  
  In documentation some disclaimer that other than GitHub, GitLab and BitBucket **MAY** not work.

</p>
</details>

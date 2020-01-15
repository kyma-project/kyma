---
title: Install components from user-defined URLs
type: Configuration
---

The Kyma Operator allows you to use external URLs as sources for the components you decide to install Kyma with. Using this mechanism, you can install Kyma with one customized component, which you, for example, modified and store in GitHub or as a `.zip` or `.tgz` archive on a server, and have other components using the officially released sources.

To install a component using an external URL as the source, you must add the **source.url** attribute to the entry of a component in the Installation custom resource (CR).

The address must expose the `chart.yaml` of the component directly. This means that for Git repositories or archives that do not store this file at the top level, you must specify the path to the file.

To specify the exact location of the `chart.yaml`, append it to the URL beginning with two backslashes `//` to indicate the path within the archive or repository. See these sample entries for components with user-defined source URLs from the Installation CR for more details:

<div tabs>
  <details>
  <summary>
  Archive URL
  </summary>

  - Archive with `chart.yaml` at the top level:
    ```
    - name: "ory"
      namespace: "kyma-system"
      source:
        url: https://hosting.com/your-user/files/kyma-custom-ory.zip
    ```

  - Archive with `chart.yaml` deeper in file structure:
    ```
    - name: "ory"
      namespace: "kyma-system"
      source:
        url: https://hosting.com/your-user/files/kyma-custom-ory.zip//kyma-custom/resources/ory
    ```

  </details>
  <details>
  <summary>
  Git repository URL
  </summary>

  >**TIP:** To get the repository URL suitable for the Installation CR, use the HTTPS address available through the GitHub web UI and remove `https://`.

  - Repository with `chart.yaml` at the top level:
    ```
    - name: "cluster-essentials"
      namespace: "kyma-system"
      source:
        url: github.com/my-project/kyma.git
    ```

  - Repository with `chart.yaml` deeper in file structure:
    ```
    - name: "cluster-essentials"
      namespace: "kyma-system"
      source:
        url: github.com/my-project/kyma.git//resources/cluster-essentials
    ```

  </details>

</div>

>**NOTE:** Read [this](#custom-resource-installation) document to learn more about the Installation CR.

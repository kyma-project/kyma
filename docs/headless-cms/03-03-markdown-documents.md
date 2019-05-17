---
title: Markdown documents
type: Details
---

The DocsTopic or ClusterDocsTopic CR supports various documentation formats, including Markdown (`.md`) documents. Each `.md` file must consist of two parts: metadata and content. The following example illustrates the required structure: 

```
---
metadata: {value or text string}
another_metadata: {value or text string}
---

{The content of your document in **Markdown**.}
```

## Metadata

Each Markdown document displayed in the Kyma Console requires metadata in a specific format called [Front Matter](https://forestry.io/docs/editing/front-matter/). 

### Structure

When you create a Markdown document, define its `title` and `type`. Place the metadata at the top of your `.md` file, and separate it with three dashes.

```
---
title: {Document title}
type: {Document type}
---
```

- The `title` metadata defines the title of your document. 
- The `type` metadata groups single documents together. Multiple documents that use the same `type` generate a grouping. For example, if you have multiple tutorials, you can group them below a navigation node called **Tutorials**.

>**NOTE:** If there is only one document of a certain type, remove the `type` metadata, so that the document displays well in the UIs.

### Display

In the Docs UI, which is the view displayed once you click on the question mark icon ![](./assets/docs-ui-question-mark.png) in the top-right corner of the Console, the metadata help create the left-side navigation structure. The Docs UI displays documents grouped under a common `type` in alphanumeric order as per files names. The following example shows four documents, their metadata, and corresponding places in the left-side navigation.

<div tabs>
  <details>
  <summary>
  Metadata source
  </summary>

```
//03-01-sidecar-proxy-injection.md
---
title: Sidecar Proxy Injection
type: Details
---
``` 
```
//03-02-istio-patch.md
---
title: Istio Patch
type: Details
---
```
```
//03-03-istio-rbac.md
---
title: Istio RBAC configuration
type: Details
---
```
```
//01-01-service-mesh.md
---
title: Overview
---
```
  </details>
  <details>
  <summary>
  Docs UI preview
  </summary>

![](./assets/left-side-nav-docs-ui.png)
  </details>
</div>

In the Service Catalog and Instances views, which contain Service Classes documentation, the `title` metadata displays as the name of a particular tab. If you don't provide `title`, the UIs display the file name as a fallback. See the following example:

<div tabs>
  <details>
  <summary>
  Metadata source
  </summary>

```
//overview.md
---
title: Overview
---
``` 
```
//plan-details.md
---
title: Services and Plans
type: Details
---
```
  </details>
  <details>
  <summary>
  Service Catalog preview
  </summary>

![](./assets/tabs-catalog-ui.png)
  </details>
</div>

>**NOTE:** A document with `title` **Overview** always displays as the first tab. Markdown files with `title` other than **Overview** appear in alphanumeric order.

## Content

Content is the body of your document. Write content in [Markdown](https://daringfireball.net/projects/markdown/syntax) which is a simplified markup language.

In Kyma, to make documentation more reader-friendly, some Markdown features are customized. See the following examples:
1. Linking - link between documents in the same topic or in different topics using metadata. 
    <div tabs>
      <details>
      <summary>
      Source
      </summary>

      Create a reference using the /{type-of-topic}/{id}#{type}-{title}-{header} pattern:
      ```
      [Kyma in a nutshell](/root/kyma/#overview-in-a-nutshell)
      ```
      </details>
      <details>
      <summary>
      Preview
      </summary>

      [Kyma in a nutshell](/root/kyma/#overview-kyma-in-a-nutshell)
      </details>
    </div>
2. Documentation toggles - render several versions of a given section in one document or have several versions of one document.
    <div tabs>
      <details>
      <summary>
      Source
      </summary>

      ```    
      <div tabs>
      <details>
      <summary>
      HTTPS
      </summary>

      `git clone https://github.com/kyma-project/kyma.git ; cd kyma`
      </details>
      <details>
      <summary>
      SSH
      </summary>

      `git clone git@github.com:kyma-project/kyma.git ; cd kyma`
      </details>
    </div>
    ```
      </details>
      <details>
      <summary>
      Preview
      </summary>

      ![](./assets/doc-toggle.png)
      </details>
    </div>
3. Panels - use colorful containers that call out important or additional information within a topic.
    <div tabs>
      <details>
      <summary>
      Source
      </summary>

      ```
      >**TIP:** If you contribute to the [bundles](https://github.com/kyma-project/bundles/tree/master/bundles) repository, you do not have to compress your bundles as the system does it automatically.
      ``` 
      </details>
      <details>
      <summary>
      Preview
      </summary>

      ![](./assets/tip-panel.png)
      </details>
    </div>

 Read the [Content Guidelines](https://github.com/kyma-project/community/tree/master/guidelines/content-guidelines) to learn more about the customized Markdown features and other rules of writing content in Kyma. 

 >**CAUTION:** Markdown customized in a different way than in Kyma may not render properly in the Console UI.
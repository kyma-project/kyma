---
title: Metadata
type: Details
---

Apart from content, each markdown document displayed in the Kyma Console requires metadata in a specific format.

## Metadata structure

When you create a document, define its `title` and `type`. Place the metadata at the top of your `.md` file, and separate it with three dashes. The following example illustrates the required structure:

```
---
title: {Document title}
type: {Document type}
---
```

- The `title` metadata defines the title of your document. 
- The `type` metadata groups single documents together. Multiple documents that use the same `type` generate a grouping. For example, if you have multiple tutorials, you can group them below a navigation node called **Tutorials**.

>**NOTE:** If there is only one document of a certain type, remove the `type` metadata, so that the document displays well in the UI.

## Metadata display

In the Docs UI, which is the view displayed once you click on the question mark icon in the top-right corner of the Console, the metadata create the left-side navigation structure. The Docs UI displays the documents grouped under a common `type` in the alphanumeric order as per files name.

In the Service Catalog and Instances views, in which you can see Service Classes documentation, the `title` metadata displays as the name of a particular tab. If you don't provide `title`, the UIs display the file name as a fallback. 

>**NOTE:** A document with `title` **Overview** always displays as the first tab. Markdown files with `title` other than **Overview** appear in alphanumeric order.

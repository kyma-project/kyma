---
title: Metadata
type: Details
---

Apart from content, each document displayed in the Kyma Console requires metadata in a specific format.

## Metadata structure

When you create a document, define its **title** and **type**. Place the metadata at the top of your file, and separate it with three dashes. The following example illustrates the required structure:

```
---
title: {Document title}
type: {Document type}
---
```

- The **title** metadata defines the title of your document. 
- The **type** metadata groups single documents together. Multiple documents that use the same **type** generate a grouping. For example, if you have multiple tutorials, you can group them below a navigation node called **Tutorials**.

## Metadata display

In the Docs UI, which is the view displayed once you click on the question mark icon in the top-right corner of the Console, the metadata create the left-side navigation structure.

In the Service Catalog and Instances views, in which you can see Service Classes documentation, the **title** metadata displays as the name of a particular tab. If you don't provide **title**, the UIs display the file name as a fallback.

---
title: Source code
type: Overview
---

Serverless in Kyma allows you to choose where you want to keep your Function's source code and dependencies. You can either place them directly in the Function CR under the **spec.source** and **spec.deps** fields as an **inline Function**, or store the code and dependencies in a public or private Git repository (**Git Functions**). Choosing the second option ensures your Function is versioned and gives you more development freedom in the choice of a project structure or an IDE.

> **TIP:** Read more about Functions of [Git source type](#details-git-source-type).

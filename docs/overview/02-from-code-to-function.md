---
title: From code to Function
type: Overview
---

Pick the programming language for the Function and decide where you want to keep the source code. Serverless will create the workload out of it for you.

## Runtimes

Functions support multiple languages by using the underlying execution environments known as runtimes. Currently, you can create both Node.js (12 & 14) and Python (3.8) Functions in Kyma.

> **TIP:** See [sample Functions](#details-sample-functions) for each available runtime.

## Source code

You can also choose where you want to keep your Function's source code and dependencies. You can either place them directly in the Function CR under the **spec.source** and **spec.deps** fields as an **inline Function**, or store the code and dependencies in a public or private Git repository (**Git Functions**). Choosing the second option ensures your Function is versioned and gives you more development freedom in the choice of a project structure or an IDE.

> **TIP:** Read more about [Git Functions](#details-git-source-type).

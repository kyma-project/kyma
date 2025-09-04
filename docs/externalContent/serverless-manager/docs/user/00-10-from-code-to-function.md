# From Code to Function

Pick the programming language for your Function and decide where you want to keep the source code. The Serverless module uses it to create the workload.

## Serverless Runtimes

Functions support programming languages by using the underlying execution environments known as runtimes. You can create both Node.js and Python Functions in Kyma.

> [!TIP]
> See [sample Functions](technical-reference/07-10-sample-functions.md) for each available runtime.

## Source Code

You can also choose where you want to keep your Function's source code and dependencies. You can either place them directly in the Function CR under the **spec.source** and **spec.deps** fields as an **inline Function**, or store the code and dependencies in a public or private Git repository (**Git Functions**). Choosing the second option ensures your Function is versioned and gives you more development freedom in the choice of a project structure or an IDE.

> [!TIP]
> Read more about [Git Functions](technical-reference/07-40-git-source-type.md).

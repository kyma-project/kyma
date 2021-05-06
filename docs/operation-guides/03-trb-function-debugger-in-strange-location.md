---
title: Function debugger stops at dependency files
type: Troubleshooting
---

If you debug your Function in `runtime=Nodejs12` or `runtime=Nodejs14` and you set a breakpoint in the first line of the main Function, the debugger can stop at dependencies.

The reason is that you shouldn't debug the first line. Add a comment in the first line instead, and start debugging from the second line.

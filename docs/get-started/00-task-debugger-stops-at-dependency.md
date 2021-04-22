---
title: Debugger stops at dependency files
type: Troubleshooting
---

<!-- I'm massively missing context here - which command did the user run, `kyma run function --debug` ? Isn't this simply a note to the topic about debugging a Function with the CLI? -->


If you debug your Function in `runtime=Nodejs12` or `runtime=Nodejs10` and you set a breakpoint in the first line of the main Function, the debugger can stop at dependencies.

The reason is that you shouldn't debug the first line. Add a comment in the first line instead, and start debugging from the second line.

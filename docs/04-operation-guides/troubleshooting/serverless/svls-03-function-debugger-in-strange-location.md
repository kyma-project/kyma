---
title: Function debugger stops at dependency files
---

<!-- I'm massively missing context here - which command did the user run, `kyma run function --debug` ? Isn't this simply a note to the topic about debugging a Function with the CLI? -->

## Symptom

If you debug your Function in `runtime=nodejs18` and you set a breakpoint in the first line of the main Function, the debugger can stop at dependencies.

## Cause

Debugging started at the first line.

## Remedy

Add a comment in the first line, and start debugging from the second line.

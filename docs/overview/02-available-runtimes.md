---
title: Available runtimes
type: Overview
---

Functions support multiple languages through the use of runtimes. To use a chosen runtime, add its name and version as a value in the **spec.runtime** field of the [Function custom resource (CR)](#custom-resource-function). If this value is not specified, it defaults to `nodejs14`. Dependencies for a Node.js Function should be specified using the [`package.json`](https://docs.npmjs.com/creating-a-package-json-file) file format. Dependencies for a Python Function should follow the format used by [pip](https://packaging.python.org/key_projects/#pip).

See [sample Functions](#details-sample-functions) for each available runtime.

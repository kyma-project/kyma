---
title: Clean up Runtime data
type: Tutorials
---

This tutorial shows how to clean up Runtime data. This operation removes a given Runtime and all its data from the database and frees up the Runtime ID for reuse. 

## Steps

> **NOTE:** To access the Runtime Provisioner, forward the port on which the GraphQL Server is listening.

To clean up Runtime data for a given Runtime, make a call to the Runtime Provisioner with a mutation like this:  

```graphql
mutation { cleanupRuntimeData(id: "61d1841b-ccb5-44ed-a9ec-45f70cd1b0d3")}
```

A successful call returns the ID of the wiped Runtime:

```graphql
{
"data": {
  "cleanupRuntimeData": "61d1841b-ccb5-44ed-a9ec-45f70cd1b0d3"
}
}
```

Use the Runtime ID to [check the Runtime Status](08-04-provisioner-runtime-status.md) and make sure it has been wiped out. 

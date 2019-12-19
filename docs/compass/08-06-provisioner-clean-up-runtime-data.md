---
title: Clean up Runtime data
type: Tutorials
---

This tutorial shows how to clean up Runtime data. This operation removes all the data for a given Runtime from the database and frees up the Runtime ID for reuse. 

## Steps

> **NOTE:** To access the Runtime Provisioner, forward the port on which the GraphQL Server is listening.

To clean up Runtime data for a given Runtime, make a call to the Runtime Provisioner with a mutation like this:  

```graphql
mutation { cleanupRuntimeData(id: "61d1841b-ccb5-44ed-a9ec-45f70cd1b0d3")
  {
    id
    message
  }
}
```

A successful call returns the Runtime ID and the message on whether the data clean-up succeeded:

```graphql
{
  "data": {
    "cleanupRuntimeData": {
      "id": "61d1841b-ccb5-44ed-a9ec-45f70cd1b0d3",
      "message": "Successfully cleaned up data for Runtime with ID 61d1841b-ccb5-44ed-a9ec-45f70cd1b0d3"
    }
  }
}
```

---
title: '"403 Forbidden" in the Console'
type: Troubleshooting
---

If you log to the Console and get the `403 Forbidden` error, do the following:

  1. Fetch the ID Token. For example, use the [Chrome Developer Tools](https://developers.google.com/web/tools/chrome-devtools) and search for the token in sent requests.
  2. Decode the ID Token. For example, use the [jwt.io](https://jwt.io/) page.
  3. Check if the token contains groups claims:
  
      ```json
      "groups": [
        "example-group"
      ]
      ```
     
  4. Make sure the group you are assigned to has [permissions](#details-roles-in-kyma) to view resources you requested.
  5. If your token does not contain group claim, check the token subject field

      ```json
      "sub": "john.doe@acme.com"
      ```
  6. Make sure user is assigned to the [permissions](#details-roles-in-kyma) to view resources you requested.
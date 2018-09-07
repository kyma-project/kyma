---
title: Manage static users in Dex
type: Details
---

Edit the `dex-config-map.yaml` ConfigMap file located in the `kyma/resources/dex/templates` directory to add, update, or remove a user in the static user store.

To add or update a user, follow this template:

```yaml
  staticPasswords:
  - email: {USER_EMAIL}
    username: {USERNAME}
    hash: {BCRYPT_HASH_OF_THE_PASSWORD}
    userID: {USER_ID}
```

This table explains the placeholders used in the template:

|Placeholder | Description |
|---|---|
| USER_EMAIL | Specifies the user's email. Should be unique among **staticPasswords** list. |
| USERNAME | Specifies the username. Should be unique among **staticPasswords** list. |
| BCRYPT_HASH_OF_THE_PASSWORD | Specifies [bcrypt hash](https://en.wikipedia.org/wiki/Bcrypt) of the user's password. |
| USER_ID | Specifies the user's identifier. Should be unique among **staticPasswords** list. |

To remove a user, delete the corresponding entry from the ConfigMap file.

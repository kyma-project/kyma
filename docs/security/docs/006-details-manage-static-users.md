---
title: Manage static users in Dex
type: Details
---

To create a static user in Dex, create a Secret with the **dex-user-config** label set to `true`. Run: 

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name:  {SECRET_NAME}
  namespace: {SECRET_NAMESPACE}
  labels:
    "dex-user-config": "true"
data:
  email: {BASE64_USER_EMAIL}
  username: {BASE64_USERNAME}
  password: {BASE64_USER_PASSWORD}  
type: Opaque
EOF
```
The following table describes the fields that are mandatory to create a static user. If any of these fields is not included, the user is not created. 

|Field | Description |
|---|---|
| data.email | Base64-encoded email address used to sign-in to the console UI. Must be unique. |
| data.username | Base64-encoded username displayed in the console UI. |
| data.password | Base64-encoded user password. There are no specific requirements regarding password strength, but it is recommended to use a password that is at least 8-characters-long. |

Create the Secrets in the cluster before Dex is installed. The Dex init-container with the tool that configures Dex generates user configuration data basing on properly labelled Secrets, and adds the data to the ConfigMap.

If you want to add a new static user after Dex is installed, restart the Dex Pod. This creates a new Pod with an updated ConfigMap.

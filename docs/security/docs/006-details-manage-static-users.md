---
title: Manage static users in Dex
type: Details
---

To create a static user create a Secret with label `"dex-user-config": "true"`.

The secret data must contain three fields, all three encoded with base64. This table contains an explaination of each required field:

|Field | Description |
|---|---|
| email | Specifies the user's email which will be used to log into the console. Should be unique. |
| username | Specifies the username which will be displayed in console. |
| password | Specifies the user's password which will be used to log into the console. There are no specific prerequisites for a password, however it is suggested to make it at least 8 characters long. |

If any of these fields will be missing the user won't be created.

> **NOTE:** The secrets should exist before the dex installation, because the init-container with the tool which configures dex runs only in the time of Dex pod creation.
> To add users after the Dex pod is created, you can restart the Dex pod. That will cause creation of another Dex pod, with configmap updated.

The admin user is created from resources/dex/templates/dex-users-secret.yaml template. You can use this template as an example how to do that properly. However, notice that the password in the template is generated with helm. To apply it manually you must change it to base64 encoded password.
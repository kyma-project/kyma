---
title: '"Not enough permissions" error'
type: Troubleshooting
---

If you log in to the Kyma Console and receive `Not enough permissions` error message:

1. Fetch the ID Token. For example, use [Chrome DevTools](https://developers.google.com/web/tools/chrome-devtools) and search for the token in the sent requests authorization header.
2. Decode the ID Token. For example, use the [jwt.io](https://jwt.io/) page.
3. Check the value of the **"email_verified"** property:

```bash
{
...
  "iss": "{your OIDC issuer}",
  "aud": [
    "kyma-client",
    "console"
  ],
  "exp": 1595525592,
  "iat": 1595496792,
  "azp": "console",
...
  "email": "{YOUR_EMAIL_ADDRESS}",
  "email_verified": false,
}
```

4. If the value is set to `false`, it means that the identity provider was unable to verify your email address. Contact your identity provider for further guidance.
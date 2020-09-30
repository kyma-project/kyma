---
title: '"Not enough permissions" error'
type: Troubleshooting
---

If you log in to the SKR and receive an error message `Not enough permissions`:

1. Fetch the ID Token. For example, use the [Chrome Developer Tools](https://developers.google.com/web/tools/chrome-devtools) and search for the token in sent requests.
2. Decode the ID Token. For example, use the [jwt.io](https://jwt.io/) page.
3. Check the value of `"email_verified"` property:

```bash
{
...
  "iss": "https://dex.c-6d073c0.kyma-stage.shoot.live.k8s-hana.ondemand.com",
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

4. If the value is `false`, make sure that you have a valid account in your IDP.
---
title: Security guidelines for MicroFrontend and ClusterMicroFrontend CRs
type: obsolete?
---

For security reasons, follow the listed guidelines when you configure the web server for the MicroFrontend or ClusterMicroFrontend:

- Make the MicroFrontend or ClusterMicroFrontend accessible only through HTTPS.
- Make the **Access-Control-Allow-Origin** HTTP header as restrictive as possible.
- Set the **X-Content-Type HTTP** header to `nosniff`.
- Set the **X-Frame-Options** HTTP header to `sameorigin` or `allow-from ALLOWED_URL`.
- Add Content Security Policies (CSPs).

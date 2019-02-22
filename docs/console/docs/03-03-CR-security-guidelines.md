---
title: Security guidelines for MicroFrontend and ClusterMicroFrontend CRs
type: Details
---

We highly recommend to include the following configuration to your `MicroFrontend` or `ClusterMicroFrontend` web server config file:
 - Make it only accessible through HTTPS.
 - Make the **Access-Control-Allow-Origin** HTTP header as restrictive as possible.
 - Set the **X-Content-Type HTTP** header to `nosniff`.
 - Set the **X-Frame-Options** HTTP header to `sameorigin` or `allow-from ALLOWED_URL`.
 - Add Content Security Policies (CSPs).

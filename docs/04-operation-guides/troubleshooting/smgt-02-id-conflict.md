---
title: '"Found more that one class" error'
---

## Symptom

ServiceClasses are visible in Service Catalog but provisioning action is blocked with the `Found more that one class` message.

## Cause

Different Service Brokers registered ClusterServiceClasses or ServiceClasses with the same ID.

## Remedy

Make sure that the addons IDs you provide in your ClusterAddonsConfiguration or AddonsConfiguration CRs are uniqe. Read about the [addons registration rules](../../03-tutorials/service-management/smgt-16-hb-register-addons-sc.md#registration-rules) for more information.

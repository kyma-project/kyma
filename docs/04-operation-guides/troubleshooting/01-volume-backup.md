---
title: Cannot create a volume snapshot
---

## Symptom

You cannot create a volume snapshot.

## Cause

If a Persistent Volume Claim is not bound to a Persistent Volume, the attempt to create a volume snapshot from that Persistent Volume Claim fails with no retries. An event will be logged to indicate no binding between the Persistent Volume Claim and the Persistent Volume.

This may happen if Persistent Volume Claim and Volume Snapshot specifications are in the same YAML file. As a result, the Volume Snapshot and the Persistent Volume Claim objects are created at the same time, but the Persistent Volume is not available yet so it cannot be bound to the Persistent Volume Claim.

## Remedy

Wait until the Persistent Volume Claim is bound to the Persistent Volume. Then, create the snapshot.

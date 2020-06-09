---
title: Cannot create a volume snapshot
type: Troubleshooting
---

If a PersistentVolumeClaim is not bound to a PersistentVolume, the attempt to create a volume snapshot from that PersistentVolumeClaim will fail with no retries. An event will be logged to indicate no binding between the PersistentVolumeClaim and the PersistentVolume.

This may happen if PersistentVolumeClaim and VolumeSnapshot specifications are in the same YAML file. As a result, the VolumeSnapshot and the PersistentVolumeClaim objects are created at the same time, but the PersisitentVolume is not available yet so it cannot be bound to the PersistentVolumeClaim. 
To solve this issue, wait until the PersistentVolumeClaim is bound to the PersistentVolume and then create the snapshot.

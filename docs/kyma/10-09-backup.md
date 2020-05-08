---
title: VolumeSnapshot creation failed
type: Troubleshooting
---

If a PersistentVolumeClaim is not bound, the attempt to create a volume snapshot from that PersistentVolumeClaim will fail with no retries. An event will be logged to indicate that the PersistentVolumeClaim is not bound.

Note that this can happen if the PersistentVolumeClaim spec and the VolumeSnapshot spec are in the same YAML file. In this case, when the VolumeSnapshot object is created, the PersistentVolumeClaim object is created but volume creation is not complete and therefore the PersistentVolumeClaim is not yet bound. You must wait until the PersistentVolumeClaim is bound and then create the snapshot.

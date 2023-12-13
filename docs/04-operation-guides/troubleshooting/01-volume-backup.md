# Cannot Create a Volume Snapshot

## Symptom

You cannot create a volume snapshot.

## Cause

If a PersistentVolumeClaim (PVC) is not bound to a PersistentVolume, the attempt to create a volume snapshot from that PVC fails with no retries. An event will be logged to indicate no binding between the PVC and the PersistentVolume.

This may happen if PVC and VolumeSnapshot specifications are in the same YAML file. As a result, the VolumeSnapshot and the PVC objects are created at the same time, but the PersistentVolume is not available yet so it cannot be bound to the PVC.

## Remedy

Wait until the PVC is bound to the PersistentVolume. Then, create the snapshot.

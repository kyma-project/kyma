---
title: Troubleshooting
---

## Events do not reach the Namespace with active Eventing

The upgrade of the Event Sources Controller Manager triggers the new Pod to update the KService for the HTTPSource. As a result of changes introduced to the KService, the Knative Serving Controller (KSC) creates a new Revision. The KSC does not pick up this new Revision and marks the KService as `Not ready`, still looking for the old Revision.
To fix this issue, delete the existing KService. It will be recreated automatically, pointing to the right Revision.

Follow these steps:

1. Check the status of the HTTP Source:

    ```bash
    $ kubectl get httpsources.sources.kyma-project.io -n kyma-integration
     NAME          READY   REASON
    application   False   ServiceNotReady
    ```

2. Check the status of the corresponding KService:

    ```bash
    $ kubectl get ksvc -n kyma-integration
    NAME          URL                                                     LATESTCREATED          LATESTREADY          READY   REASON
    application   http://application.kyma-integration.svc.cluster.local   application-g4qd8      application-c2zlz    False   RevisionMissing
    ```

3. If the PrinterColumn `READY` displays the value `False`, and `REASON` shows `RevisionMissing`, delete the existing KService:

    ```bas
    $ kubectl delete ksvc -n kyma-integration application
    ```

4. The KService will recreate automatically, pointing to the right revision. Check the status of the KService and the HTTP Source:

    ```bash
    $ kubectl get ksvc -n kyma-integration application
    NAME          URL                                                     LATESTCREATED          LATESTREADY          READY   REASON
    application   http://application.kyma-integration.svc.cluster.local   application-w57fv      application-w57fv    True
    
    $ kubectl get httpsources.sources.kyma-project.io -n kyma-integration application
    NAME          READY   REASON
    application   True
    ```
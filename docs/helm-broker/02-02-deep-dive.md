---
title: Architecture deep dive
type: Architecture
---

The diagram and steps describe the Helm Broker workflow in details, including the logic of its inner components, namely the Controller and Broker.

![Architecture deep dive](./assets/hb-deep-dive.svg)

1. The Controller watches for ClusterAddonsConfiguration (CAC) and AddonsConfiguration (AC) custom resources (CRs).
2. The user creates, updates, or deletes CAC or AC custom resources.
3. The Controller fetches and parses the data of all addon repositories defined in these custom resources. During this step, the Controller does the following:
  - Analyze fetched addons against errors.
  - Check for ID duplications under the **repositories** field.
  - Check for ID conflicts with already registered addons.
4. The Controller persists fetched addons in the storage.
5. When the first CR appears, the Controller creates ClusterServiceBroker or ServiceBroker, depending on the type of the CR. The ClusterServiceBroker/ServiceBroker provides information about Broker's proper endpoint to the Service Catalog. This endpoint returns the list of available services. There is always only one ClusterServiceBroker and one ServiceBroker per Namespace, no matter the number of CRs.
6. The Broker component reads addons from the storage and exposes them as Service Classes to the Service Catalog.
7. The Service Catalog communicates with ClusterServiceBroker/ServiceBroker and watches for Service Classes.

### Update CRs

There are two cases in which you might want to update your CR:
- Re-fetching addons from a remote server
- Changing repositories URLs

If you provided changes to your addon on a remote server but the URL did not change, you must re-fetch your changes manually. In such a case, increment the **reprocessRequest** field to explicitly request the reprocessing of already registered and processed CR.

If you made any change in your addon's URLs, the update process is triggered automatically and the Controller performs its logic.

### Delete CRs

If you want to delete a given CR, follow the clean-up logic:

1. If a given CR is in the **Ready** state, remove it from the storage.
2. After all CRs with the **Ready** status are removed, increment the **reprocessRequest** field of all failed custom resources in order to reprocess them.

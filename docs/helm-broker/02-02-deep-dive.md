---
title: Architecture deep dive
type: Architecture
---

The diagram and steps describe the Helm Broker workflow in details, including the logic of its inner components, namely the Controller and Broker.

![Architecture deep dive](./assets/hb-deep-dive.svg)

1. The Controller watches for ClusterAddonsConfiguration (CAC) and AddonsConfiguration (AC) custom resources (CRs).
2. The user creates, updates, or deletes CAC or AC custom resources.
3. The Controller fetches and parses the data of all addon repositories defined under the **spec.repositories** field in a given CR. During this step, the Controller does the following:
  - Analyze fetched addons against errors.
  - Check for ID duplications under the **repositories** field.
  - Check for ID conflicts with already registered addons.
4. The Controller persists fetched addons in the storage.
5. When the first CR appears, the Controller creates ClusterServiceBroker or ServiceBroker, depending on the type of the CR. The ClusterServiceBroker/ServiceBroker provides information about Broker's proper endpoint to the Service Catalog. This endpoint returns the list of available services. There is always only one ClusterServiceBroker per cluster and one ServiceBroker per Namespace, no matter the number of existing CRs. Whenever the list of offered services changes, the Controller triggers the Service Catalog to fetch services from the ClusterServiceBroker/ServiceBroker.
6. The Broker component fetches addons from the storage and exposes them to the Service Catalog.
7. The Service Catalog calls the catalog endpoint of the ClusterServiceBroker/ServiceBroker and creates the Service Classes.

## Update CRs

There are two cases in which you might want to update your CR:
- Re-fetching addons from a remote server
- Changing repositories URLs

### Re-fetching addons

If you provided changes to your addon on a remote server but the URL did not change, you must re-fetch your changes manually. In such a case, increment the **reprocessRequest** field to explicitly request the reprocessing of already registered and processed CR.

### Changing repositories URLs

If you made any change in your addon's URLs, the update process is triggered automatically and the Controller performs its logic.

## Delete CRs

This is the logic the Controller executes after you delete a given CR:

1. If a given CR is in the **Ready** state, the Controller removes it from the storage.
2. After addons are removed from the storage, the Controller increments the **reprocessRequest** field of all failed CRs that had conflicts with the deleted CR in order to reprocess them.
3. The Controller deletes a finalizer from the given CR.

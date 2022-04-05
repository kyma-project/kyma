const {assert} = require('chai');
const k8s = require('@kubernetes/client-node');
const {sleep, getVirtualService} = require('../utils');
const {
  queryLoki,
  lokiSecretData,
  tryGetLokiPersistentVolumClaim,
  getVirtualServices,
} = require('./client');

async function checkLokiLogs(startTimestamp) {
  const labels = '{app="commerce-mock", container="mock", namespace="mocks"}';
  let logsFetched = false;
  let retries = 0;
  while (retries < 20) {
    const logs = await queryLoki(labels, startTimestamp);
    if (logs.streams.length > 0) {
      logsFetched = true;
      break;
    }
    await sleep(5 * 1000);
    retries++;
  }
  assert.isTrue(logsFetched, 'No logs fetched from Loki');
}

async function checkLokiLogsInKymaNamespaces(startTimestamp) {
  const labels = ['{namespace="kyma-system"}',
    '{namespace="kyma-integration"}'];
  let logsFetched = false;
  let retries = 0;
  for (let el = 0; el < labels.length; el++) {
    while (retries < 20) {
      logsFetched = false;
      const logs = await queryLoki(labels[el], startTimestamp);
      if (logs.streams.length > 0) {
        logsFetched = true;
        break;
      }
      await sleep(5 * 1000);
      retries++;
    }
  }
  assert.isTrue(logsFetched, 'No logs fetched from Loki');
}

async function checkRetentionPeriod() {
  const secretData = k8s.loadYaml(await lokiSecretData());

  assert.equal(secretData?.table_manager?.retention_period, '120h');
  assert.equal(secretData?.chunk_store_config?.max_look_back_period, '120h');
}

async function checkPersistentVolumeClaimSize() {
  const pvc = await tryGetLokiPersistentVolumClaim();
  if (pvc == null) {
    console.log('Loki PVC not found. Skipping...');
    return;
  }
  assert.equal(pvc.status.capacity.storage, '30Gi');
}

async function checkIfLokiVirutalServiceIsPresence() {
  const hosts = getVirtualService('kyma-system', 'loki');
  // const hosts = getVirtualService('kyma-system', 'monitoring-grafana');
  console.log('hosts', hosts);
  assert.isEmpty(hosts, 'Loki is exposed via Virtual Service');
}

async function checkVirtualServicePresence() {
  const virtualServices = await getVirtualServices();
  let lokiVSPresence = false;
  for (let index = 0; index < virtualServices.length; index++) {
    if (virtualServices[index]?.metadata?.name === 'loki') {
      lokiVSPresence = true;
      break;
    }
  }

  assert.isFalse(lokiVSPresence, 'Loki is exposed via Virtual Service');
}

module.exports = {
  checkLokiLogs,
  checkLokiLogsAllNamespaces: checkLokiLogsInKymaNamespaces,
  checkRetentionPeriod,
  checkIfLokiVirutalServiceIsPresence,
  checkPersistentVolumeClaimSize,
  checkVirtualServicePresence,
};

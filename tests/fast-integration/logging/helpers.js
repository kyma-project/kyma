const {assert} = require('chai');
const k8s = require('@kubernetes/client-node');
const {sleep} = require('../utils');
const {
  queryLoki,
  lokiSecretData,
  lokiPersistentVolumClaim,
  getVirtualServices,
} = require('./client');

// checkLokiLogs used directly in Commerce Mock tests.
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
    await sleep(5*1000);
    retries++;
  }
  assert.isTrue(logsFetched, 'No logs fetched from Loki');
}

// Required checks have been set as per https://github.com/kyma-project/kyma/issues/11136.
async function checkRetentionPeriod() {
  const secretData = k8s.loadYaml(await lokiSecretData());
  let periodCheck = false;
  if (secretData?.chunk_store_config?.max_look_back_period == '120h' &&
  secretData?.table_manager?.retention_period == '120h') {
    periodCheck = true;
  }
  assert.isTrue(periodCheck, 'Loki retention_period or max_look_back_period is not 120h');
}

async function checkPersistentVolumeClaimSize() {
  const pvc = await lokiPersistentVolumClaim();
  let claimSizeCheck = false;
  if (pvc.status.capacity.storage == '30Gi') {
    claimSizeCheck = true;
  }
  assert.isTrue(claimSizeCheck, 'Claim size for loki is not 30Gi');
}

async function checkVirtualServicePresence() {
  const virtualServices = await getVirtualServices();
  let lokiVSPresence = false;
  for (let index = 0; index < virtualServices.length; index++) {
    if (virtualServices[index]?.metadata?.name == 'loki') {
      lokiVSPresence = virtualServices[index]?.metadata?.name == 'loki';
      break;
    }
  }
  assert.isFalse(lokiVSPresence, 'Loki is exposed via Virtual Service');
}

module.exports = {
  checkLokiLogs,
  checkRetentionPeriod,
  checkPersistentVolumeClaimSize,
  checkVirtualServicePresence,
};

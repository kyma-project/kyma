const {assert} = require('chai');
const k8s = require('@kubernetes/client-node');
const {
  lokiSecretData,
  tryGetLokiPersistentVolumeClaim,
  logsPresentInLoki,
} = require('./client');

async function checkCommerceMockLogsInLoki(startTimestamp) {
  const labels = '{app="commerce-mock", container="mock", namespace="mocks"}';

  const commerceMockLogsPresent = await logsPresentInLoki(labels, startTimestamp);

// checkLokiLogs used directly in Commerce Mock tests.
async function checkLokiLogs(startTimestamp, labels) {
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
  assert.isTrue(commerceMockLogsPresent, 'No logs from commerce mock present in Loki');
}

async function checkKymaLogsInLoki(startTimestamp) {
  const systemLabel = '{namespace="kyma-system"}';
  const integrationLabel = '{namespace="kyma-integration"}';

  const kymaSystemLogsPresent = await logsPresentInLoki(systemLabel, startTimestamp);
  const kymaIntegrationLogsPresent = await logsPresentInLoki(integrationLabel, startTimestamp);

  assert.isTrue(kymaSystemLogsPresent, 'No logs from kyma-system namespace present in Loki');
  assert.isTrue(kymaIntegrationLogsPresent, 'No logs from kyma-integration namespace present in Loki');
}

async function checkRetentionPeriod() {
  const secretData = k8s.loadYaml(await lokiSecretData());

  assert.equal(secretData?.table_manager?.retention_period, '120h');
  assert.equal(secretData?.chunk_store_config?.max_look_back_period, '120h');
}

async function checkPersistentVolumeClaimSize() {
  const pvc = await tryGetLokiPersistentVolumeClaim();
  if (pvc == null) {
    console.log('Loki PVC not found. Skipping...');
    return;
  }

  assert.equal(pvc.status.capacity.storage, '30Gi');
}

module.exports = {
  checkCommerceMockLogsInLoki,
  checkKymaLogsInLoki,
  checkRetentionPeriod,
  checkPersistentVolumeClaimSize,
};

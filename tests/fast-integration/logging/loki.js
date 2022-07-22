const {assert} = require('chai');
const k8s = require('@kubernetes/client-node');
const {
  lokiSecretData,
  tryGetLokiPersistentVolumeClaim,
  logsPresentInLoki,
} = require('./client');
const {info} = require('../utils');

async function checkCommerceMockLogs(startTimestamp) {
  const labels = '{app="commerce-mock", container="mock", namespace="mocks"}';

  const commerceMockLogsPresent = await logsPresentInLoki(labels, startTimestamp);

  assert.isTrue(commerceMockLogsPresent, 'No logs from commerce mock present in Loki');
}

async function checkKymaLogs(startTimestamp) {
  const systemLabel = '{namespace="kyma-system"}';

  const kymaSystemLogsPresent = await logsPresentInLoki(systemLabel, startTimestamp);

  assert.isTrue(kymaSystemLogsPresent, 'No logs from kyma-system namespace present in Loki');
}

async function checkRetentionPeriod() {
  const secretData = k8s.loadYaml(await lokiSecretData());

  assert.equal(secretData?.table_manager?.retention_period, '120h');
  assert.equal(secretData?.chunk_store_config?.max_look_back_period, '120h');
}

async function checkPersistentVolumeClaimSize() {
  const pvc = await tryGetLokiPersistentVolumeClaim();
  if (pvc == null) {
    info('Loki PVC not found. Skipping...');
    return;
  }

  assert.equal(pvc.status.capacity.storage, '30Gi');
}

module.exports = {
  checkCommerceMockLogs,
  checkKymaLogs,
  checkRetentionPeriod,
  checkPersistentVolumeClaimSize,
};

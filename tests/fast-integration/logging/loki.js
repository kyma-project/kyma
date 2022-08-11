const {assert} = require('chai');
const k8s = require('@kubernetes/client-node');
const {
  lokiSecretData,
  tryGetLokiPersistentVolumeClaim,
  logsPresentInLoki,
  queryLoki,
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

function parseJson(str) {
  try {
    return JSON.parse(str);
  } catch (e) {
    return undefined;
  }
}

async function verifyIstioAccessLogFormat(startTimestamp) {
  const query = '{container="istio-proxy",namespace="kyma-system",pod="logging-loki-0"}';

  const responseBody = await queryLoki(query, startTimestamp);

  assert.isTrue(responseBody.data.result[0].values.length > 0, 'No Istio access logs found for loki');
  const entry = JSON.parse(responseBody.data.result[0].values[0][1]);
  const log = parseJson(entry.log);
  assert.isDefined(log, `Istio access log is not in JSON format: ${entry.log}`);
  assert.isDefined(log['response_code'], `Istio access log does not have 'response_code' field: ${log}`);
  assert.isDefined(log['bytes_received'], `Istio access log does not have 'bytes_received' field: ${log}`);
  assert.isDefined(log['bytes_sent'], `Istio access log does not have 'bytes_sent' field: ${log}`);
  assert.isDefined(log['duration'], `Istio access log does not have 'duration' field: ${log}`);
  assert.isDefined(log['start_time'], `Istio access log does not have 'start_time' field: ${log}`);
}

module.exports = {
  checkCommerceMockLogs,
  checkKymaLogs,
  checkRetentionPeriod,
  checkPersistentVolumeClaimSize,
  verifyIstioAccessLogFormat,
};

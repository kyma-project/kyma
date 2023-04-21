module.exports = {
  checkCommerceMockLogs,
  checkKymaLogs,
  checkRetentionPeriod,
  checkPersistentVolumeClaimSize,
  verifyIstioAccessLogFormat,
};

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
  // Iterate over the values
  const numberOfLogs = responseBody.data.result[0].values.length;
  let entry;
  let log;
  for (let i =0; i<= numberOfLogs; i++) {
    // Some logs dont have values[i][1]. In such a case skip the log line
    if (typeof responseBody.data.result[0].values[i][1] === 'undefined') {
      continue;
    }
    entry = JSON.parse(responseBody.data.result[0].values[i][1]);
    log = parseJson(entry.log);
    if (isJsonString(entry.log) ) {
      verifyLogAttributeIsPresent('method', log);
      verifyLogAttributeIsPresent('path', log);
      verifyLogAttributeIsPresent('protocol', log);
      verifyLogAttributeIsPresent('response_code', log);
      verifyLogAttributeIsPresent('response_flags', log);
      verifyLogAttributeIsPresent('response_code_details', log);
      verifyLogAttributeIsPresent('connection_termination_details', log);
      verifyLogAttributeIsPresent('upstream_transport_failure_reason', log);
      verifyLogAttributeIsPresent('bytes_received', log);
      verifyLogAttributeIsPresent('bytes_sent', log);
      verifyLogAttributeIsPresent('duration', log);
      verifyLogAttributeIsPresent('upstream_service_time', log);
      verifyLogAttributeIsPresent('x_forwarded_for', log);
      verifyLogAttributeIsPresent('user_agent', log);
      verifyLogAttributeIsPresent('request_id', log);
      verifyLogAttributeIsPresent('authority', log);
      verifyLogAttributeIsPresent('upstream_host', log);
      verifyLogAttributeIsPresent('upstream_cluster', log);
      verifyLogAttributeIsPresent('upstream_local_address', log);
      verifyLogAttributeIsPresent('downstream_local_address', log);
      verifyLogAttributeIsPresent('downstream_remote_address', log);
      verifyLogAttributeIsPresent('requested_server_name', log);
      verifyLogAttributeIsPresent('route_name', log);
      verifyLogAttributeIsPresent('traceparent', log);
      verifyLogAttributeIsPresent('tracestate', log);
      return;
    }
  }
  log = parseJson(entry.log);
  assert.isDefined(log, `Istio access log is not in JSON format: ${entry.log}`);
}

function verifyLogAttributeIsPresent(attribute, logBody) {
  assert.isDefined(logBody[attribute], `Istio access log does not have '${attribute}' field: ${logBody}`);
}

function isJsonString(str) {
  try {
    JSON.parse(str);
  } catch (e) {
    return false;
  }
  return true;
}

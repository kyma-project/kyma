module.exports = {
  checkCommerceMockLogs,
  checkKymaLogs,
  checkFluentBitLogs,
  checkRetentionPeriod,
  checkPersistentVolumeClaimSize,
  verifyIstioAccessLogFormat,
};

const {assert} = require('chai');
const k8s = require('@kubernetes/client-node');
const {
  lokiConfigData,
  tryGetLokiPersistentVolumeClaim,
  logsPresentInLoki,
  queryLoki,
} = require('./client');
const {
  info,
  debug,
} = require('../utils');

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

async function checkFluentBitLogs(startTimestamp) {
  const labels = '{container="fluent-bit", namespace="kyma-system"}';

  const fluentBitLogsPresent = await logsPresentInLoki(labels, startTimestamp, 1);

  assert.isFalse(fluentBitLogsPresent, 'Fluent Bit logs present in Loki');
}

async function checkRetentionPeriod() {
  const secretData = k8s.loadYaml(await lokiConfigData());

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

async function verifyIstioAccessLogFormat(startTimestamp) {
  const query = '{container="istio-proxy",namespace="kyma-system",pod="logging-loki-test-0"}';

  const accessLogsPresent = await logsPresentInLoki(query, startTimestamp);
  assert.isTrue(accessLogsPresent, 'No Istio access logs present in Loki');

  const responseBody = await queryLoki(query, startTimestamp);
  assert.isDefined(responseBody.data.result[0].values, 'Empty response for the query for Istio access logs');
  assert.isTrue(responseBody.data.result[0].values.length > 0, 'No Istio access logs found for loki');
  const numberOfResults = responseBody.data.result.length;
  // Iterate over the values
  for (let i = 0; i <= numberOfResults; i++) {
    const result = responseBody.data.result[i];
    if (accessLogVerified(result)) {
      return;
    }
  }
  assert.fail('Istio access log is not present: ', JSON.stringify(responseBody.data));
}

function accessLogVerified(result) {
  const numberOfLogs = result.values.length;
  for (let i =0; i<= numberOfLogs; i++) {
    // Some logs dont have values[i][1]. In such a case skip the log line
    const val = result.values[i];
    if ( !Array.isArray(val) ) {
      debug('skipping while its not an array', JSON.stringify(val));
      continue;
    }
    if (val.length < 2) {
      debug('skipping length not > 1: ', JSON.stringify(val[1]));
      continue;
    }
    if (isJsonString(val[1])) {
      const log = JSON.parse(val[1]);
      if (typeof log['method'] === 'undefined') {
        debug('skipping while method is not present', JSON.stringify(log));
        continue;
      }
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
      return true;
    }
  }
  return false;
}

function verifyLogAttributeIsPresent(attribute, logBody) {
  assert.isDefined(logBody[attribute],
      `Istio access log does not have '${attribute}' field: ${JSON.stringify(logBody)}`);
}

function isJsonString(str) {
  try {
    JSON.parse(str);
  } catch (e) {
    return false;
  }
  return true;
}

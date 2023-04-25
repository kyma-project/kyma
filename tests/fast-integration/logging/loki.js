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
const {
  info,
  sleep,
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
  await sleep(10*1000);
  const query = '{container="istio-proxy",namespace="kyma-system",pod="logging-loki-0"}';
  const responseBody = await queryLoki(query, startTimestamp);
  console.log('responseBody', JSON.stringify(responseBody));
  numberOfResults = responseBody.data.result.length;
  assert.isDefined(responseBody.data.result[0].values, 'Empty response for the query for Istio access logs');
  assert.isTrue(responseBody.data.result[0].values.length > 0, 'No Istio access logs found for loki');
  console.log('number of results', numberOfResults);
  // Iterate over the values
  for (let i = 0; i <= numberOfResults; i++) {
    const result = responseBody.data.result[i];
    // console.log(numberOfLogs);
    // console.log(responseBody.data.result[0].values);
    if (accessLogVerified(result)) {
      return;
    }
  }
  assert.throws(JSON.stringify(responseBody.data), `Istio access log is not present`);
}

function accessLogVerified(result) {
  const numberOfLogs = result.values.length;
  for (let i =0; i<= numberOfLogs; i++) {
    // Some logs dont have values[i][1]. In such a case skip the log line
    console.log(Array.isArray(result.values[i]));
    const val = result.values[i];
    if ( !Array.isArray(val) ) {
      console.log('skipping while its not an array' + val + '\n');
      continue;
    }
    if (val.length < 2) {
      console.log('skipping length not > 1: ' + val[1] + '\n');
      continue;
    }
    if (isJsonString(val[1])) {
      log = JSON.parse(val[1]);
      if (typeof log['method'] === 'undefined') {
        console.log('skipping while method is not present' + JSON.stringify(log) + '\n');
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

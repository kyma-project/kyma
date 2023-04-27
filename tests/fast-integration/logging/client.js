module.exports = {
  logsPresentInLoki,
  tryGetLokiPersistentVolumeClaim,
  lokiSecretData,
  queryLoki,
  createIstioAccessLogResource,
  loadResourceFromFile,
};

const {
  convertAxiosError,
  getPersistentVolumeClaim,
  getSecretData,
  sleep,
  k8sApply,
} = require('../utils');
const {proxyGrafanaDatasource} = require('../monitoring/client');

const fs = require('fs');
const path = require('path');
const k8s = require('@kubernetes/client-node');

async function logsPresentInLoki(query, startTimestamp, iterations = 20) {
  for (let i = 0; i < iterations; i++) {
    const responseBody = await queryLoki(query, startTimestamp);
    if (responseBody.data.result.length > 0) {
      return true;
    }
    await sleep(5 * 1000);
  }
  return false;
}

async function tryGetLokiPersistentVolumeClaim() {
  try {
    return await getPersistentVolumeClaim('kyma-system', 'storage-logging-loki-0');
  } catch (err) {
    return null;
  }
}

async function lokiSecretData() {
  const secretData = await getSecretData('logging-loki', 'kyma-system');
  return secretData['loki.yaml'];
}

async function queryLoki(query, startTimestamp) {
  const path = `loki/api/v1/query_range?query=${query}&start=${startTimestamp}`;
  try {
    const response = await proxyGrafanaDatasource('Loki', path, 5, 30, 10000);
    return response.data;
  } catch (err) {
    throw convertAxiosError(err, 'cannot query loki');
  }
}


async function createIstioAccessLogResource() {
  const istioAccessLogsResource = loadResourceFromFile('./istio_access_logs.yaml');
  const namespace = 'kyma-system';
  await k8sApply(istioAccessLogsResource, namespace);
}

function loadResourceFromFile(file) {
  const yaml = fs.readFileSync(path.join(__dirname, file), {
    encoding: 'utf8',
  });
  return k8s.loadAllYaml(yaml);
}

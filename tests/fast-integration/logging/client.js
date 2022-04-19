const axios = require('axios');
const {
  convertAxiosError,
  getPersistentVolumeClaim,
  getSecretData,
  sleep,
  retryPromise,
  info,
} = require('../utils');
const {getGrafanaUrl} = require('../monitoring/client');


async function getGrafanaDatasourceId(grafanaUrl, datasourceName) {
  const url = `${grafanaUrl}/api/datasources/id/${datasourceName}`;

  return retryPromise(async () => await axios.get(url), 5, 1000);
}

async function getLokiViaGrafana(path, retries = 5, interval = 30, timeout = 10000) {
  const grafanaUrl = await getGrafanaUrl();
  const lokiDatasourceResponse = await getGrafanaDatasourceId(grafanaUrl, 'Loki');
  const lokiDatasourceId = lokiDatasourceResponse.data.id;
  const url = `${grafanaUrl}/api/datasources/proxy/${lokiDatasourceId}/loki/${path}`;
  info('loki grafana url', url);

  return retryPromise(async () => await axios.get(url, {timeout: timeout}),
      retries, interval);
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

async function logsPresentInLoki(query, startTimestamp) {
  for (let i = 0; i < 20; i++) {
    const responseBody = await queryLoki(query, startTimestamp);
    if (responseBody.data.result.length > 0) {
      return true;
    }
    await sleep(5 * 1000);
  }
  return false;
}

async function queryLoki(query, startTimestamp) {
  const path = `api/v1/query_range?query=${query}&start=${startTimestamp}`;
  try {
    const response = await getLokiViaGrafana(path);
    return response.data;
  } catch (err) {
    throw convertAxiosError(err, 'cannot query loki');
  }
}

module.exports = {
  logsPresentInLoki,
  tryGetLokiPersistentVolumeClaim,
  lokiSecretData,
};

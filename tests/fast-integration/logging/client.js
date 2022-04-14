const axios = require('axios');
const https = require('https');
const {
  convertAxiosError,
  getPersistentVolumeClaim,
  getSecretData,
  sleep,
  retryPromise,
  info,
} = require('../utils');
const {getGrafanaUrl} = require('../monitoring/client');

async function getLokiViaGrafana(path, retries = 5, interval = 30, timeout = 10000) {
  const grafanaUrl = await getGrafanaUrl();
  const url = `${grafanaUrl}/api/datasources/proxy/3/loki/${path}`;
  info('loki grafana url', url);
  delete axios.defaults.headers.common['Accept'];
  const httpsAgent = new https.Agent({
    rejectUnauthorized: false,
  });

  return retryPromise(async () => await axios.get(url, {httpsAgent: httpsAgent, timeout: timeout}),
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
    const logs = await queryLoki(query, startTimestamp);
    if (logs.streams.length > 0) {
      return true;
    }
    await sleep(5 * 1000);
  }
  return false;
}

async function queryLoki(query, startTimestamp) {
  const path = `api/prom/query?query=${query}&start=${startTimestamp}`;
  try {
    const responseBody = await getLokiViaGrafana(path);
    info('responseBody', responseBody);
    return responseBody.data;
  } catch (err) {
    throw convertAxiosError(err, 'cannot query loki');
  }
}

module.exports = {
  logsPresentInLoki,
  tryGetLokiPersistentVolumeClaim,
  lokiSecretData,
};

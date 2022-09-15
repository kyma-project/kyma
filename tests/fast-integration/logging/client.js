const {
  convertAxiosError,
  getPersistentVolumeClaim,
  getSecretData,
  sleep,
} = require('../utils');
const {proxyGrafanaDatasource} = require('../monitoring/client');


async function getLokiViaGrafana(path, retries = 5, interval = 30, timeout = 10000) {
  return await proxyGrafanaDatasource('Loki', path, retries, interval, timeout);
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

async function queryLoki(query, startTimestamp) {
  const path = `loki/api/v1/query_range?query=${query}&start=${startTimestamp}`;
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
  queryLoki,
};

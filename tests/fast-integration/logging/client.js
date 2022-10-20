module.exports = {
  logsPresentInLoki,
  tryGetLokiPersistentVolumeClaim,
  lokiSecretData,
  queryLoki,
};

const {
  convertAxiosError,
  getPersistentVolumeClaim,
  getSecretData,
  sleep,
} = require('../utils');
const {proxyGrafanaDatasource} = require('../monitoring/client');

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

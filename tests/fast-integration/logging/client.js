const axios = require('axios');
const {
  kubectlPortForward,
  retryPromise,
  convertAxiosError,
  getPersistentVolumeClaim,
  getSecretData,
  sleep,
} = require('../utils');

const lokiPort = 3100;

function lokiPortForward() {
  return kubectlPortForward('kyma-system', 'logging-loki-0', lokiPort);
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

async function queryLoki(labels, startTimestamp) {
  const url = `http://localhost:${lokiPort}/api/prom/query?query=${labels}&start=${startTimestamp}`;
  try {
    const responseBody = await retryPromise(() => axios.get(url, {timeout: 10000}), 5);
    return responseBody.data;
  } catch (err) {
    throw convertAxiosError(err, 'cannot query loki');
  }
}

module.exports = {
  lokiPortForward,
  queryLoki,
  logsPresentInLoki,
  tryGetLokiPersistentVolumeClaim,
  lokiSecretData,
};

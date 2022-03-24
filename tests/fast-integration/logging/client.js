
const axios = require('axios');
const {
  kubectlPortForward,
  retryPromise,
  convertAxiosError,
  getPersistentVolumeClaim,
  getSecretData,
  getAllVirtualServices,
} = require('../utils');

const lokiPort = 3100;

function lokiPortForward() {
  return kubectlPortForward('kyma-system', 'logging-loki-0', lokiPort);
}

async function lokiPersistentVolumClaim() {
  return await getPersistentVolumeClaim('kyma-system', 'storage-logging-loki-0');
}

async function lokiSecretData() {
  const secretData = await getSecretData('logging-loki', 'kyma-system');
  return secretData['loki.yaml'];
}

function getVirtualServices() {
  return getAllVirtualServices();
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
  getVirtualServices,
  lokiPersistentVolumClaim,
  lokiSecretData,
};

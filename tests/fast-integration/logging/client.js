const {
  convertAxiosError,
  getPersistentVolumeClaim,
  getSecretData,
  sleep,
  callServiceViaProxy,
} = require('../utils');

function getLoki(path) {
  return callServiceViaProxy('kyma-system', 'logging-loki', '3100', path);
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
  const path = `api/prom/query?query=${query}&start=${startTimestamp}`;
  try {
    const responseBody = await getLoki(path);
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


const axios = require('axios');
const {
  kubectlPortForward,
  retryPromise,
  convertAxiosError,
} = require('../utils');

const lokiPort = 3100;

function lokiPortForward() {
  return kubectlPortForward('kyma-system', 'logging-loki-0', lokiPort);
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
};

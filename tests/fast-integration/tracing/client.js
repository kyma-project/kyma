const {
  debug,
  convertAxiosError,
} = require('../utils');
const {proxyGrafanaDatasource} = require('../monitoring/client');


async function getJaegerViaGrafana(path, retries = 5, interval = 30,
    timeout = 50000, debugMsg = undefined) {
  return await proxyGrafanaDatasource('Jaeger', path, retries, interval, timeout, debugMsg);
}

async function getJaegerTrace(traceId) {
  const path = `api/traces/${traceId}`;

  debug(`fetching trace: ${traceId} from jaeger`);

  try {
    const debugMsg = `waiting for trace (id: ${traceId}) from jaeger...`;
    const responseBody = await getJaegerViaGrafana(path, 30, 1000, 30 * 1000, debugMsg);
    return responseBody.data;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get jaeger trace');
  }
}

module.exports = {
  getJaegerTrace,
};
